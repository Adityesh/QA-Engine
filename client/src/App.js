import logo from './logo.svg';
import './App.css';
import {useEffect} from 'react';

function App() {
  useEffect(() => {
    fetch('/')
    .then(data => data.text())
    .then(res => console.log(res))
  })

  return (
    <div className="App">
      
    </div>
  );
}

export default App;
