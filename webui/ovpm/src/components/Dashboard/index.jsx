import React from "react";

import { GetAuthToken, ClearAuthToken } from "../../utils/auth.js";
import { API } from "../../utils/restClient.js";
import { baseURL, endpoints } from "../../api.js";
import UserDashboard from "./UserDashboard";
import AdminDashboard from "./AdminDashboard";

export default class Dashboard extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      username: "",
      isAdmin: false
    };
    let authToken = GetAuthToken();
    this.api = new API(baseURL, endpoints, authToken);
  }

  handleGetUserSuccess(res) {
    this.setState({
      username: res.data.user.username,
      isAdmin: res.data.user.is_admin
    });
  }

  componentWillMount() {
    this.api.call(
      "authStatus",
      {},
      true,
      this.handleGetUserSuccess.bind(this),
      this.handleGetUserFailure.bind(this)
    );
  }

  handleGetUserFailure(error) {
    if (error.response.status === 401) {
      this.handleAuthFailure(error);
    }

    console.log("get user failure", error);
  }

  handleAuthFailure(error) {
    console.log("auth failure", error);
    ClearAuthToken();
  }

  render() {
    let dashboard;
    if (!this.state.isAdmin) {
      dashboard = (
        <UserDashboard username={this.state.username} api={this.api} />
      );
    } else {
      dashboard = (
        <AdminDashboard username={this.state.username} api={this.api} />
      );
    }
    return <div className="mui--text-center">{dashboard}</div>;
  }
}
