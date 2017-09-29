//export const baseURL = "http://192.168.14.200:9091/v1"//ev
//export const baseURL = "http://172.16.16.149:9091/v1" //ofislaptop
//export const baseURL = "http://172.16.16.79:9091/v1"  //ofis desktop
export var baseURL = window.location.protocol + "//" + window.location.host + "/api/v1"

if (process.env.NODE_ENV !== 'production') {
    baseURL = "http://172.16.16.149:9091/api/v1"
}

export const endpoints = {
    authenticate: {
        path: "/auth/authenticate",
        method: "POST",
    },
    authStatus: {
        path: "/auth/status",
        method: "GET",
    },
    genConfig: {
        path: "/user/genconfig",
        method: "POST",
    },
    userList: {
        path: "/user/list",
        method: "GET",
    },
    userCreate: {
        path: "/user/create",
        method: "POST",
    },
    userDelete: {
        path: "/user/delete",
        method: "POST",
    },
    userUpdate: {
        path: "/user/update",
        method: "POST",
    },
    networkList: {
        path: "/network/list",
        method: "GET",
    },
    vpnStatus: {
        path: "/vpn/status",
        method: "GET",
    },
    netDefine: {
        path: "/network/create",
        method: "POST",
    },
    netUndefine: {
        path: "/network/delete",
        method: "POST",
    },
    netAssociate: {
        path: "/network/associate",
        method: "POST",
    },
    netDissociate: {
        path: "/network/dissociate",
        method: "POST",
    }
}
