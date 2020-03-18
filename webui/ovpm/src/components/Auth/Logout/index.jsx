import React from "react";
import { Redirect } from "react-router";
import { ClearAuthToken } from "../../../utils/auth.js";
export default class Logout extends React.Component {
  componentWillMount() {
    ClearAuthToken(); // Logout
  }

  render() {
    return <Redirect to="/login" />;
  }
}
