import React from 'react';
import ReactDOM from 'react-dom';
import { Router  } from 'react-router-dom'
import Master from './components/Master'

const App = () => (
    <div>
    <Master />
    </div>
)

ReactDOM.render(
  <App />, document.getElementById('root')
);
