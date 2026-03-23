import { useGameStore } from '../../hooks/useGameStore.js';

const API_BASE = '';
const wsProto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${wsProto}//${window.location.host}/api/ws`;

let _ws = null;
let _lastSeenSeq = 0;

export const client = {
  async start() {
    const res = await fetch(`${API_BASE}/api/engine/start`, { method: 'POST' });
    useGameStore.getState().setRunning(true);
    this.connectWS();
    return res.json();
  },

  async pause() {
    const res = await fetch(`${API_BASE}/api/engine/pause`, { method: 'POST' });
    useGameStore.getState().setRunning(false);
    this.disconnectWS();
    return res.json();
  },

  connectWS() {
    if (_ws) return;
    const ws = new WebSocket(WS_URL);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data);
        const store = useGameStore.getState();

        if (data.world) {
          store.setWorld(data.world);
        }

        if (data.events) {
          const events = data.events;
          const totalCount = data.totalCount || 0;
          if (totalCount > _lastSeenSeq) {
            const newCount = totalCount - _lastSeenSeq;
            const newEvents = events.slice(Math.max(0, events.length - newCount));
            _lastSeenSeq = totalCount;
            store.appendEvents(newEvents);
          }
        }

        if (data.stats) {
          store.setRunning(data.stats.running);
          if (data.stats.tickCount != null) {
            store.setTickCount(data.stats.tickCount);
          }
        }
      } catch (err) {
        console.warn('[client] ws message error:', err);
      }
    };

    ws.onclose = () => {
      _ws = null;
    };

    ws.onerror = (err) => {
      console.warn('[client] ws error:', err);
      ws.close();
    };

    _ws = ws;
  },

  disconnectWS() {
    if (_ws) {
      _ws.close();
      _ws = null;
    }
  },

  async waitForServer() {
    // Poll /health until the server reports ready (genesis complete)
    const store = useGameStore.getState();
    while (true) {
      try {
        const res = await fetch(`${API_BASE}/health`);
        const data = await res.json();
        if (res.ok && data.status === 'ok') return;
        // Update genesis progress in store
        if (data.progress) {
          store.setGenesisProgress(data.progress);
        }
      } catch (_) { /* server not yet available */ }
      await new Promise(r => setTimeout(r, 1500));
    }
  },

  async fetchInitialWorld() {
    // Wait until genesis is complete before fetching world data
    await this.waitForServer();
    useGameStore.getState().setGenesisComplete(true);

    try {
      const [worldRes, eventsRes, statsRes] = await Promise.all([
        fetch(`${API_BASE}/api/world`),
        fetch(`${API_BASE}/api/events`),
        fetch(`${API_BASE}/api/stats`),
      ]);
      const store = useGameStore.getState();
      let world = null;
      if (worldRes.ok) {
        world = await worldRes.json();
        store.setWorld(world);
      }
      if (eventsRes.ok) {
        const data = await eventsRes.json();
        const events = data.events || [];
        const totalCount = data.totalCount || 0;
        if (events.length > 0) {
          store.appendEvents(events);
        }
        _lastSeenSeq = totalCount;
      }
      if (statsRes.ok) {
        const stats = await statsRes.json();
        store.setRunning(stats.running);
        if (stats.tickCount != null) {
          store.setTickCount(stats.tickCount);
        }
        if (stats.running) {
          this.connectWS();
        }
      }
      return world;
    } catch (_) { /* fetch failed after readiness — unexpected */ }
    return null;
  },
};
