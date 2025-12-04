const { createApp } = Vue;

createApp({
  data() {
    return {
      view: "landing",          // 'landing' | 'auth' | 'dashboard'
      authMode: "login",        // 'login' | 'register'
      authForm: {
        email: "",
        password: ""
      },
      authLoading: false,
      authError: "",
      user: null,
      token: null,

      // Dados principais
      items: [],       // informações + memórias
      guardians: [],   // pessoas de confiança
      settings: {
        emergency_protocol_enabled: false
      },

      // Composer
      composerType: "info", // 'info' | 'guardian' | 'memory'
      infoForm: {
        title: "",
        content: ""
      },
      guardianForm: {
        name: "",
        email: ""
      },
      memoryForm: {
        title: "",
        content: ""
      },
      savingInfo: false,
      savingGuardian: false,
      savingMemory: false,

      // Feed
      feedFilter: "all",  // 'all' | 'info' | 'guardian' | 'memory'

      // Assistente
      assistantInput: "",
      assistantReply: "",
      assistantLoading: false,

      // Modais
      showLgpd: false,
      showSettings: false
    };
  },

  computed: {
    summaryCounts() {
      const infos = this.items.filter(i => i.type === "info").length;
      const memories = this.items.filter(i => i.type === "memory").length;
      const guardians = this.guardians.length;
      return {
        infos,
        memories,
        guardians,
        total: infos + memories + guardians
      };
    },

    unifiedEntries() {
      const itemEntries = this.items.map(it => ({
        id: it.id,
        kind: it.type === "memory" ? "memory" : "info",
        typeLabel: it.type === "memory" ? "Memória / mensagem" : "Informação importante",
        title: it.title || "(sem título)",
        description: it.content || "",
        updatedAt: it.updatedAt || it.createdAt || null
      }));

      const guardianEntries = this.guardians.map(g => ({
        id: g.id,
        kind: "guardian",
        typeLabel: "Pessoa de confiança",
        title: g.name || g.email || "Pessoa de confiança",
        description: g.email || "",
        updatedAt: g.updatedAt || g.createdAt || null
      }));

      return [...itemEntries, ...guardianEntries].sort((a, b) => {
        const da = a.updatedAt ? new Date(a.updatedAt).getTime() : 0;
        const db = b.updatedAt ? new Date(b.updatedAt).getTime() : 0;
        return db - da;
      });
    },

    filteredEntries() {
      if (this.feedFilter === "all") {
        return this.unifiedEntries;
      }
      return this.unifiedEntries.filter(e => e.kind === this.feedFilter);
    }
  },

  methods: {
    /* Navegação básica */

    goToAuth(mode) {
      this.authMode = mode || "login";
      this.view = "auth";
      this.authError = "";
    },

    goToLanding() {
      this.view = "landing";
      this.authError = "";
    },

    goToDashboard() {
      this.view = "dashboard";
      this.fetchAllData();
    },

    /* Auth */

    async handleAuthSubmit() {
      this.authError = "";
      if (!this.authForm.email || !this.authForm.password) {
        this.authError = "Preencha e-mail e senha.";
        return;
      }
      this.authLoading = true;
      try {
        const endpoint =
            this.authMode === "login" ? "/api/auth/login" : "/api/auth/register";

        const res = await fetch(endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(this.authForm)
        });

        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          this.authError =
              body.message || "Não foi possível autenticar. Tente novamente.";
          return;
        }

        const body = await res.json();
        // Assumindo que venha { token, user: { email, ... } }
        this.token = body.token;
        this.user = body.user || { email: this.authForm.email };

        this.authForm.password = "";
        this.goToDashboard();
      } catch (err) {
        console.error(err);
        this.authError = "Erro de conexão. Tente novamente em instantes.";
      } finally {
        this.authLoading = false;
      }
    },

    logout() {
      this.token = null;
      this.user = null;
      this.items = [];
      this.guardians = [];
      this.view = "landing";
    },

    /* Fetch helpers */

    authHeaders() {
      const headers = { "Content-Type": "application/json" };
      if (this.token) {
        headers["Authorization"] = `Bearer ${this.token}`;
      }
      return headers;
    },

    async fetchAllData() {
      await Promise.all([this.fetchItems(), this.fetchGuardians(), this.fetchSettings()]);
    },

    async fetchItems() {
      try {
        const res = await fetch("/api/items", {
          headers: this.authHeaders()
        });
        if (!res.ok) return;
        const data = await res.json();
        this.items = Array.isArray(data) ? data : data.items || [];
      } catch (e) {
        console.error("Erro ao buscar items:", e);
      }
    },

    async fetchGuardians() {
      try {
        const res = await fetch("/api/guardians", {
          headers: this.authHeaders()
        });
        if (!res.ok) return;
        const data = await res.json();
        this.guardians = Array.isArray(data) ? data : data.guardians || [];
      } catch (e) {
        console.error("Erro ao buscar guardians:", e);
      }
    },

    async fetchSettings() {
      try {
        const res = await fetch("/api/settings", {
          headers: this.authHeaders()
        });
        if (!res.ok) return;
        const data = await res.json();
        this.settings.emergency_protocol_enabled =
            data.emergency_protocol_enabled ?? false;
      } catch (e) {
        console.error("Erro ao buscar settings:", e);
      }
    },

    /* Composer actions */

    async saveInfo() {
      if (!this.infoForm.title || !this.infoForm.content) return;
      this.savingInfo = true;
      try {
        const payload = {
          title: this.infoForm.title,
          type: "info",
          content: this.infoForm.content
        };
        const res = await fetch("/api/items", {
          method: "POST",
          headers: this.authHeaders(),
          body: JSON.stringify(payload)
        });
        if (!res.ok) return;
        const created = await res.json();
        this.items.unshift(created);
        this.infoForm.title = "";
        this.infoForm.content = "";
        this.feedFilter = "all";
      } catch (e) {
        console.error("Erro ao salvar info:", e);
      } finally {
        this.savingInfo = false;
      }
    },

    async saveGuardian() {
      if (!this.guardianForm.name || !this.guardianForm.email) return;
      this.savingGuardian = true;
      try {
        const payload = {
          name: this.guardianForm.name,
          email: this.guardianForm.email
        };
        const res = await fetch("/api/guardians", {
          method: "POST",
          headers: this.authHeaders(),
          body: JSON.stringify(payload)
        });
        if (!res.ok) return;
        const created = await res.json();
        this.guardians.unshift(created);
        this.guardianForm.name = "";
        this.guardianForm.email = "";
        this.feedFilter = "guardian";
      } catch (e) {
        console.error("Erro ao salvar guardian:", e);
      } finally {
        this.savingGuardian = false;
      }
    },

    async saveMemory() {
      if (!this.memoryForm.title || !this.memoryForm.content) return;
      this.savingMemory = true;
      try {
        const payload = {
          title: this.memoryForm.title,
          type: "memory",
          content: this.memoryForm.content
        };
        const res = await fetch("/api/items", {
          method: "POST",
          headers: this.authHeaders(),
          body: JSON.stringify(payload)
        });
        if (!res.ok) return;
        const created = await res.json();
        this.items.unshift(created);
        this.memoryForm.title = "";
        this.memoryForm.content = "";
        this.feedFilter = "memory";
      } catch (e) {
        console.error("Erro ao salvar memória:", e);
      } finally {
        this.savingMemory = false;
      }
    },

    /* Feed actions */

    formatDate(dt) {
      if (!dt) return "";
      const d = new Date(dt);
      if (Number.isNaN(d.getTime())) return "";
      return d.toLocaleDateString("pt-BR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric"
      });
    },

    formattedEntryDate(entry) {
      // Tenta usar updatedAt; se não tiver, usa createdAt; se nada der certo, retorna vazio
      const dt = entry.updatedAt || entry.createdAt;
      return this.formatDate(dt);
    },

    editEntry(entry) {
      // MVP: apenas joga os dados de volta para o composer,
      // e muda o tipo atual. Edição real pode ser feita depois.
      if (entry.kind === "guardian") {
        this.composerType = "guardian";
        this.guardianForm.name = entry.title;
        this.guardianForm.email = entry.description;
      } else if (entry.kind === "info") {
        this.composerType = "info";
        this.infoForm.title = entry.title;
        this.infoForm.content = entry.description;
      } else if (entry.kind === "memory") {
        this.composerType = "memory";
        this.memoryForm.title = entry.title;
        this.memoryForm.content = entry.description;
      }
      window.scrollTo({ top: 0, behavior: "smooth" });
    },

    async deleteEntry(entry) {
      const ok = window.confirm("Tem certeza que deseja excluir este registro?");
      if (!ok) return;

      try {
        if (entry.kind === "guardian") {
          await fetch(`/api/guardians/${entry.id}`, {
            method: "DELETE",
            headers: this.authHeaders()
          });
          this.guardians = this.guardians.filter(g => g.id !== entry.id);
        } else {
          await fetch(`/api/items/${entry.id}`, {
            method: "DELETE",
            headers: this.authHeaders()
          });
          this.items = this.items.filter(i => i.id !== entry.id);
        }
      } catch (e) {
        console.error("Erro ao excluir entry:", e);
      }
    },

    /* Settings */

    openSettings() {
      this.showSettings = true;
    },

    async saveSettings() {
      try {
        await fetch("/api/settings", {
          method: "POST",
          headers: this.authHeaders(),
          body: JSON.stringify(this.settings)
        });
      } catch (e) {
        console.error("Erro ao salvar settings:", e);
      }
    },

    /* Assistente */

    async sendAssistant() {
      if (!this.assistantInput.trim()) return;
      this.assistantLoading = true;
      this.assistantReply = "";
      try {
        const res = await fetch("/api/assistant", {
          method: "POST",
          headers: this.authHeaders(),
          body: JSON.stringify({ input: this.assistantInput })
        });
        if (!res.ok) {
          this.assistantReply =
              "Não consegui responder agora. Tente novamente em alguns instantes.";
          return;
        }
        const data = await res.json();
        this.assistantReply =
            data.reply ||
            data.message ||
            "Entendi. Você pode usar o espaço acima para guardar essa informação na sua Caixa Famli.";
        this.assistantInput = "";
      } catch (e) {
        console.error("Erro assistant:", e);
        this.assistantReply =
            "Não consegui responder agora. Tente novamente em alguns instantes.";
      } finally {
        this.assistantLoading = false;
      }
    }
  }
}).mount("#app");
