import React, { useEffect, useRef } from 'react';
import { useGameStore } from './hooks/useGameStore';
import { useIsMobile } from './hooks/useIsMobile';
import { createGame, destroyGame } from './game/PhaserGame';
import { client } from './game/systems/client';
import StatusOverlay from './ui/StatusOverlay';
import EventLog from './ui/EventLog';
import Inspector from './ui/Inspector';
import Minimap from './ui/Minimap';
import BottomToolbar from './ui/BottomToolbar';
import LoadingScreen from './ui/LoadingScreen';
import WorldLoadingSpinner from './ui/WorldLoadingSpinner';
import RelationshipsPopup from './ui/popups/RelationshipsPopup';
import FactionsPopup from './ui/popups/FactionsPopup';
import EconomyPopup from './ui/popups/EconomyPopup';
import DemographicsPopup from './ui/popups/DemographicsPopup';
import CreationsPopup from './ui/popups/CreationsPopup';
import ActionsPopup from './ui/popups/ActionsPopup';
import LeaderboardPopup from './ui/popups/LeaderboardPopup';
import TimelinePopup from './ui/popups/TimelinePopup';
import PrayerPopup from './ui/popups/PrayerPopup';
import CombatPopup from './ui/popups/CombatPopup';
import WealthPopup from './ui/popups/WealthPopup';
import MoodPopup from './ui/popups/MoodPopup';
import WorldPopup from './ui/popups/WorldPopup';
import ChroniclePopup from './ui/popups/ChroniclePopup';
import './App.css';

function PhaserCanvas() {
  const containerRef = useRef(null);
  const gameRef = useRef(null);

  useEffect(() => {
    if (containerRef.current && !gameRef.current) {
      gameRef.current = createGame(containerRef.current);
    }
    return () => {
      destroyGame();
      gameRef.current = null;
    };
  }, []);

  return <div ref={containerRef} className="phaser-canvas" />;
}

const POPUP_MAP = {
  world: WorldPopup,
  relationships: RelationshipsPopup,
  factions: FactionsPopup,
  economy: EconomyPopup,
  demographics: DemographicsPopup,
  creations: CreationsPopup,
  actions: ActionsPopup,
  leaderboard: LeaderboardPopup,
  timeline: TimelinePopup,
  prayer: PrayerPopup,
  combat: CombatPopup,
  wealth: WealthPopup,
  mood: MoodPopup,
  chronicle: ChroniclePopup,
};

export default function App() {
  useIsMobile();
  const worldReady = useGameStore(s => s.worldReady);
  const genesisComplete = useGameStore(s => s.genesisComplete);
  const sceneReady = useGameStore(s => s.sceneReady);
  const inspectorOpen = useGameStore(s => s.inspectorOpen);
  const activePopup = useGameStore(s => s.activePopup);
  const isMobile = useGameStore(s => s.isMobile);
  const toggleChronicle = useGameStore(s => s.toggleChronicle);

  useEffect(() => {
    client.fetchInitialWorld();
    useGameStore.getState().enterWorld();

    // Handle URL-based NPC selection: /npc/<id>?myprophet=true
    const match = window.location.pathname.match(/^\/npc\/([^/]+)/);
    if (match) {
      const npcId = match[1];
      const params = new URLSearchParams(window.location.search);
      const store = useGameStore.getState();
      store.selectNpc(npcId);
      store.setFollowNpc(npcId);
      if (params.get('myprophet') === 'true') {
        store.setMyAgent(npcId);
      }
      // Clean query params but keep the /npc/<id> path
      if (window.location.search) {
        window.history.replaceState(null, '', `/npc/${npcId}`);
      }
    }

    return () => client.disconnectWS();
  }, []);

  useEffect(() => {
    const handleKey = (e) => {
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
      switch (e.key.toLowerCase()) {
        case 'm':
          useGameStore.getState().toggleMinimap();
          break;
        case 'escape':
          useGameStore.getState().closeAllPanels();
          useGameStore.getState().clearSelection();
          break;
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, []);

  const PopupComponent = activePopup ? POPUP_MAP[activePopup] : null;

  return (
    <div className="app-root">
      {!worldReady && !genesisComplete && <LoadingScreen />}
      {!worldReady && genesisComplete && <WorldLoadingSpinner />}
      {worldReady && !sceneReady && <WorldLoadingSpinner />}
      {worldReady && <>
        <div className="main-column">
          <PhaserCanvas />
          {sceneReady && <BottomToolbar />}
        </div>
        {sceneReady && <EventLog />}
        {sceneReady && <StatusOverlay />}
        {sceneReady && isMobile && (
          <button className="sdv-btn chronicle-toggle" onClick={toggleChronicle} title="Chronicle">
            📖
          </button>
        )}
        {sceneReady && <Minimap />}
        {sceneReady && inspectorOpen && <Inspector />}
        {sceneReady && PopupComponent && <PopupComponent />}
      </>}
    </div>
  );
}
