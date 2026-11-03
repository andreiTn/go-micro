(() => {
  const cfg = window.__CONFIG__ || { brokerUrl: "http://localhost:8080" };
  const brokerUrl = (cfg.brokerUrl || "http://localhost:8080").replace(/\/$/, "");

  const els = {
    actionSeg: document.getElementById("actionSeg"),
    transportSeg: document.getElementById("transportSeg"),
    transportField: document.getElementById("transportField"),
    payload: document.getElementById("payload"),
    sendBtn: document.getElementById("sendBtn"),
    resetBtn: document.getElementById("resetBtn"),
    statusLine: document.getElementById("statusLine"),
    sent: document.getElementById("sent"),
    received: document.getElementById("received"),
    brokerLabel: document.getElementById("brokerLabel"),
  };

  let action = "ping";
  let transport = "rpc";

  const defaults = {
    ping: { note: "Broker health ping — empty POST body" },
    auth: {
      action: "auth",
      auth: { email: "admin@example.com", password: "verysecret" },
    },
    log: {
      action: "log",
      transport: "rpc",
      log: { name: "console", data: "hello from API console" },
    },
    mail: {
      action: "mail",
      mail: {
        from: "console@go-micro.local",
        to: "demo@example.com",
        subject: "Go Micro mail test",
        message: "Sent from the API Console",
      },
    },
  };

  function activeButton(seg, attr, value) {
    [...seg.querySelectorAll(".seg-btn")].forEach((btn) => {
      btn.classList.toggle("is-active", btn.dataset[attr] === value);
    });
  }

  function pretty(value) {
    return JSON.stringify(value, null, 2);
  }

  function resetPayload() {
    if (action === "ping") {
      els.payload.value = pretty(defaults.ping);
      return;
    }
    if (action === "log") {
      const body = structuredClone(defaults.log);
      body.transport = transport;
      els.payload.value = pretty(body);
      return;
    }
    els.payload.value = pretty(defaults[action]);
  }

  function setStatus(text, ok) {
    els.statusLine.textContent = text;
    els.statusLine.classList.toggle("is-ok", ok === true);
    els.statusLine.classList.toggle("is-err", ok === false);
  }

  async function callBroker(path, body) {
    const started = performance.now();
    const res = await fetch(`${brokerUrl}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: body == null ? "{}" : JSON.stringify(body),
    });
    const text = await res.text();
    let data;
    try {
      data = JSON.parse(text);
    } catch {
      data = { raw: text };
    }
    const ms = Math.round(performance.now() - started);
    return { res, data, ms };
  }

  els.actionSeg.addEventListener("click", (e) => {
    const btn = e.target.closest("[data-action]");
    if (!btn) return;
    action = btn.dataset.action;
    activeButton(els.actionSeg, "action", action);
    els.transportField.hidden = action !== "log";
    resetPayload();
  });

  els.transportSeg.addEventListener("click", (e) => {
    const btn = e.target.closest("[data-transport]");
    if (!btn) return;
    transport = btn.dataset.transport;
    activeButton(els.transportSeg, "transport", transport);
    if (action === "log") resetPayload();
  });

  els.resetBtn.addEventListener("click", resetPayload);

  els.sendBtn.addEventListener("click", async () => {
    els.sendBtn.classList.add("is-pending");
    setStatus("Sending…");

    try {
      let path = "/handle";
      let body;

      if (action === "ping") {
        path = "/";
        body = {};
        els.sent.textContent = pretty({ method: "POST", path: "/" });
      } else {
        body = JSON.parse(els.payload.value);
        if (action === "log") {
          body.transport = body.transport || transport;
        }
        els.sent.textContent = pretty(body);
      }

      // Force a reflow so the fade animation restarts.
      els.received.replaceWith(els.received.cloneNode(true));
      const received = document.getElementById("received");

      const { res, data, ms } = await callBroker(path, body);
      received.textContent = pretty(data);
      const ok = res.ok;
      setStatus(`${res.status} ${res.statusText} · ${ms}ms`, ok);
    } catch (err) {
      document.getElementById("received").textContent = String(err);
      setStatus(String(err.message || err), false);
    } finally {
      els.sendBtn.classList.remove("is-pending");
    }
  });

  if (els.brokerLabel) els.brokerLabel.textContent = brokerUrl;
  resetPayload();
})();
