 const STEP = 10;
 const NOSTEP = 0
export const ws = 'ws://localhost:8080/ws'
export const playerMoves = {
  "ArrowRight": { y: `${NOSTEP}`, x: `${STEP}` },
  "ArrowLeft": { y: `${NOSTEP}`, x: `-${STEP}` },
  "ArrowUp": { y: `${STEP}`, x: `${NOSTEP}` },
  "ArrowDown": { y: `-${STEP}`, x: `${NOSTEP}` }
}