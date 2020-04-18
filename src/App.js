
import React, { useCallback } from 'react';
import './App.css';
import { useWebSocket } from './hooks/useWebSocket'
import { useBodyBounderies } from './hooks/useBodyBounderies'
import {ws, playerMoves} from './Utils.js/Utils'
import { useEventListener } from './hooks/useEventListener'
import Monkeys from './components/Monkeys'

function App() {
  const bodyBounderies = useBodyBounderies()
  const [playState, sendMSG] = useWebSocket(ws, bodyBounderies)
  const showKeyCode = useCallback(({ key }) => {
      sendMSG(playerMoves[key])
    },[sendMSG]);

  useEventListener('keydown', showKeyCode);
  return (
    <>
    <div>
      <h2>Score</h2>
      {playState.players.map(player => <div key={player.id}>{player.active? "You" : player.exceptionType} - {player.score}</div>)}
    </div>
    {playState.exceptions.map(exception => <div key={Math.random()} style={{ position: 'absolute',bottom: exception.y + 'px', left: exception.x + 'px' }}>{exception.exceptionType}</div>)}

    <Monkeys players={playState.players} />
    </>
  );
}
export default App;