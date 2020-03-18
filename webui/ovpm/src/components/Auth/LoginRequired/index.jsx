import React from "react";

import { Redirect } from "react-router";

import { IsAuthenticated } from "../../../utils/auth.js";

export default function loginRequired(WrappedComponent) {
  return class extends React.Component {
    componentWillReceiveProps(nextProps) {
      console.log("Current props: ", this.props);
      console.log("Next props: ", nextProps);
      this.setState({ isLoggedIn: false });
    }
    componentWillMount() {
      this.setState({ isLoggedIn: IsAuthenticated() });
    }
    render() {
      if (!this.state.isLoggedIn) {
        return <Redirect to="/login" />;
      }
      return <WrappedComponent {...this.props} />;
    }
  };
}
