export var baseURL =
  window.location.protocol + "//" + window.location.host + "/api/v1";

if (process.env.NODE_ENV !== "production") {
  // baseURL = "http://172.16.16.53:8080/api/v1" // local pc external ip
  baseURL = "http://127.0.0.1:8080/api/v1";
}

export const endpoints = {
  authenticate: {
    path: "/auth/authenticate",

    method: "POST"
  },
  authStatus: {
    path: "/auth/status",
    method: "GET"
  },
  genConfig: {
    path: "/user/genconfig",
    method: "POST"
  },
  userList: {
    path: "/user/list",
    method: "GET"
  },
  userCreate: {
    path: "/user/create",
    method: "POST"
  },
  userDelete: {
    path: "/user/delete",
    method: "POST"
  },
  userUpdate: {
    path: "/user/update",
    method: "POST"
  },
  networkList: {
    path: "/network/list",
    method: "GET"
  },
  vpnStatus: {
    path: "/vpn/status",
    method: "GET"
  },
  vpnRestart: {
    path: "/vpn/restart",
    method: "POST"
  },
  netDefine: {
    path: "/network/create",
    method: "POST"
  },
  netUndefine: {
    path: "/network/delete",
    method: "POST"
  },
  netAssociate: {
    path: "/network/associate",
    method: "POST"
  },
  netDissociate: {
    path: "/network/dissociate",
    method: "POST"
  }
};
