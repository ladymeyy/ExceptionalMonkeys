
    import React, { useEffect, useState, useRef } from 'react';
    import './App.css';
    import monkey from './monkey.png'
    import banana from './banana.png'

    
    
    const socket = new WebSocket('ws://localhost:8080/ws');
    
    
    
    
    function App() {
    
      const [players, setPlayers] = useState(["roei","mey"]);
      const [location, settLocation] = useState([
         {
          y: "0", x: "0"
         }
        ]
      );

      // const isFirstRun = useRef(true);
    
      function showKeyCode(e) {
        //right
        if (e.keyCode === 39) {
          socket.send(JSON.stringify({y:"0",x:"10"}))
        }
        //left
        if (e.keyCode === 37) {
          socket.send(JSON.stringify({y:"0",x:"-10"}))
        }
    
    
      }
      useEffect(() => {
        document.body.addEventListener('keyup', showKeyCode);
      }, [])
    
    

    
    
      useEffect(() => {
    
        socket.onmessage = (event) => {
           const parseData = JSON.parse(event.data)
           console.log("TCL: socket.onmessage -> parseData", parseData.p)
        
         
          
          settLocation([parseData.p])

        }
        socket.onopen = () => {
          console.log("connected successfuly")
        }
    
        socket.onclose = (e) => {
          console.log("socket close connection", e)
        }
    
    
        socket.onerror = (e) => {
          console.log("socket error", e)
        }
      }, []);
    
    
      return (
        <>
         
          <img className="monkey"  style={{ left: location[0].x + 'px' }} src={monkey}  />
    
        </>
      );
    }
    
    export default App;
    

