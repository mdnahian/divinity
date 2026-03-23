import { create } from 'zustand';

export const useGameStore = create((set, get) => ({
  world: null,
  worldReady: false,
  genesisProgress: null,
  genesisComplete: false,
  sceneReady: false,
  events: [],
  running: false,
  tickCount: 0,

  enteredWorld: false,
  myAgentNpcId: localStorage.getItem('divinity_myAgent') || null,

  selectedNpc: null,
  selectedLocation: null,
  followNpc: null,

  inspectorOpen: false,
  minimapVisible: true,
  activePopup: null,
  isMobile: false,
  chronicleOpen: false,

  setWorld: (world) => set({ world, worldReady: true }),
  setWorldReady: (worldReady) => set({ worldReady }),
  setGenesisProgress: (genesisProgress) => set({ genesisProgress }),
  setGenesisComplete: (genesisComplete) => set({ genesisComplete }),
  setSceneReady: (sceneReady) => set({ sceneReady }),
  setRunning: (running) => set({ running }),
  setTickCount: (tickCount) => set({ tickCount }),

  enterWorld: () => set({ enteredWorld: true }),
  setMyAgent: (npcId) => {
    if (npcId) localStorage.setItem('divinity_myAgent', npcId);
    else localStorage.removeItem('divinity_myAgent');
    set({ myAgentNpcId: npcId });
  },

  appendEvent: (entry) => set((state) => {
    const next = [...state.events, entry];
    if (next.length > 400) next.splice(0, next.length - 400);
    return { events: next };
  }),
  appendEvents: (entries) => set((state) => {
    if (!entries || entries.length === 0) return state;
    const next = [...state.events, ...entries];
    if (next.length > 400) next.splice(0, next.length - 400);
    return { events: next };
  }),

  selectNpc: (npcId) => {
    set({ selectedNpc: npcId, selectedLocation: null, inspectorOpen: true });
    if (npcId) window.history.pushState(null, '', `/npc/${npcId}`);
  },
  selectLocation: (locId) => set({ selectedLocation: locId, selectedNpc: null, inspectorOpen: true }),
  clearSelection: () => {
    set({ selectedNpc: null, selectedLocation: null, followNpc: null });
    if (window.location.pathname.startsWith('/npc/')) {
      window.history.pushState(null, '', '/');
    }
  },
  setFollowNpc: (npcId) => set({ followNpc: npcId }),

  toggleInspector: () => set((s) => ({ inspectorOpen: !s.inspectorOpen })),
  toggleMinimap: () => set((s) => ({ minimapVisible: !s.minimapVisible })),
  openPopup: (name) => set({ activePopup: name }),
  closePopup: () => set({ activePopup: null }),
  togglePopup: (name) => set((s) => ({ activePopup: s.activePopup === name ? null : name })),
  setIsMobile: (isMobile) => set({ isMobile }),
  toggleChronicle: () => set((s) => ({ chronicleOpen: !s.chronicleOpen })),
  closeChronicle: () => set({ chronicleOpen: false }),
  closeAllPanels: () => set({ inspectorOpen: false, activePopup: null, chronicleOpen: false }),

  phaserScene: null,
  setPhaserScene: (scene) => set({ phaserScene: scene }),
}));
