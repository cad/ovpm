import React from "react";
import "./style.css";

import { BrowserRouter as Router, Route, Redirect } from "react-router-dom";

import Login from "../Auth/Login";
import Logout from "../Auth/Logout";
import LoginRequired from "../Auth/LoginRequired";

import Dashboard from "../Dashboard";

function Home(props) {
  return <Redirect to="/dashboard" />;
}

const Master = props => (
  <Router>
    <div>
      <Route path="/" component={Home} />
      <Route path="/login" component={Login} />
      <Route path="/logout" component={Logout} />
      <Route path="/dashboard" component={LoginRequired(Dashboard)} />
    </div>
  </Router>
);
export default Master;
