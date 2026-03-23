import React, { useState, useEffect } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';
import './ChroniclePopup.css';

export default function ChroniclePopup() {
  const world = useGameStore(s => s.world);
  const chronicles = world?.chronicles || [];
  const [page, setPage] = useState(Math.max(0, chronicles.length - 1));

  useEffect(() => {
    setPage(Math.max(0, chronicles.length - 1));
  }, [chronicles.length]);

  if (chronicles.length === 0) {
    return (
      <PopupModal icon="📖" title="The Divine Chronicle">
        <div className="chronicle-empty">
          <p>The pages are blank. The chronicle awaits the passage of days...</p>
        </div>
      </PopupModal>
    );
  }

  const entry = chronicles[page];

  return (
    <PopupModal icon="📖" title="The Divine Chronicle">
      <div className="chronicle-book">
        <div className={`chronicle-page ${page}`}>
          <div className="chronicle-day">Day {entry?.day}</div>
          <h3 className="chronicle-title">{entry?.title}</h3>
          <div className="chronicle-text">
            {entry?.text?.split('\n').map((para, i) => (
              para.trim() ? <p key={i}>{para}</p> : null
            ))}
          </div>
        </div>

        <div className="chronicle-nav">
          <button
            className="sdv-btn chronicle-nav-btn"
            disabled={page <= 0}
            onClick={() => setPage(page - 1)}
          >
            ◀ Previous
          </button>
          <span className="chronicle-page-num">
            Page {page + 1} of {chronicles.length}
          </span>
          <button
            className="sdv-btn chronicle-nav-btn"
            disabled={page >= chronicles.length - 1}
            onClick={() => setPage(page + 1)}
          >
            Next ▶
          </button>
        </div>
      </div>
    </PopupModal>
  );
}
