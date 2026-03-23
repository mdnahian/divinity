import React from 'react';
import ReactDOM from 'react-dom';
import { useGameStore } from '../hooks/useGameStore';
import './PopupModal.css';

export default function PopupModal({ title, icon, children }) {
  const closePopup = useGameStore(s => s.closePopup);

  return ReactDOM.createPortal(
    <div className="popup-overlay" onClick={closePopup} onPointerDown={e => e.stopPropagation()} onWheel={e => e.stopPropagation()}>
      <div className="popup-modal sdv-frame sdv-frame-bottom" onClick={e => e.stopPropagation()} onPointerDown={e => e.stopPropagation()}>
        <div className="popup-header">
          <h3 className="popup-title">
            {icon && <span className="popup-title-icon">{icon}</span>}
            <span className="popup-title-text">{title}</span>
          </h3>
          <button className="sdv-btn panel-close" onClick={closePopup}>✕</button>
        </div>
        <div className="popup-body sdv-scroll">
          {children}
        </div>
      </div>
    </div>,
    document.body
  );
}
