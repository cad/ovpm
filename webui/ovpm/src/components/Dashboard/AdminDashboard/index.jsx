import React from "react";

import {ClearAuthToken, GetAuthToken} from "../../../utils/auth.js";
import {API} from "../../../utils/restClient.js";
import {baseURL, endpoints} from "../../../api.js";
import {Redirect} from "react-router";
import Panel from "muicss/lib/react/panel";
import Button from "muicss/lib/react/button";
import Container from "muicss/lib/react/container";
import Tabs from "muicss/lib/react/tabs";
import Tab from "muicss/lib/react/tab";
import Modal from "react-modal";
import UserEdit from "./UserEdit";
import NetworkEdit from "./NetworkEdit";
import UserPicker from "./UserPicker";
import NetworkMonitor from "./Components/NetworkMonitor";
import SystemMonitor from "./Components/SystemMonitor";

import moment from "moment";

const modalStyle = {
    content: {
        width: "50%",
        height: "600px",
        marginLeft: "auto",
        marginRight: "auto",
        padding: 0
    }
};
const CREATINGNEWUSER = "CREATINGNEWUSER";
const EDITINGUSER = "EDITINGUSER";
const DEFININGNEWNETWORK = "DEFININGNEWNETWORK";
const ASSOCIATINGUSER = "ASSOCIATINGUSER";
const DISSOCIATINGUSER = "DISSOCIATINGUSER";

let saveData = (function () {
    var a = document.createElement("a");
    document.body.appendChild(a);
    a.style = "display: none";
    return function (data, fileName) {
        var json = data,
            blob = new Blob([json], {type: "octet/stream"}),
            url = window.URL.createObjectURL(blob);
        a.href = url;
        a.download = fileName;
        a.click();
        window.URL.revokeObjectURL(url);
    };
})();

function dot2num(dot) {
    var d = dot.split(".");
    return ((+d[0] * 256 + +d[1]) * 256 + +d[2]) * 256 + +d[3];
}

function num2dot(num) {
    var d = num % 256;
    for (let i = 3; i > 0; i--) {
        num = Math.floor(num / 256);
        d = (num % 256) + "." + d;
    }
    return d;
}

export default class AdminDashboard extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            logout: false,
            users: [],
            networks: [],
            vpn: {},
            statistics: [],
            modal: "",
            self: {},
            editedUser: {},
            genConfigUsername: "",
            assocNetworkName: "",
            dissocNetworkName: "",
            possibleAssocUsers: [],
            possibleDissocUsers: [],
            autoRefreshEnabled: true,
            refreshInterval: 5000,
            systemStatus: null,
            networkInterfaces: [],
            selectedInterface: '',
            interfaceStats: null
        };
        let authToken = GetAuthToken();
        this.api = new API(baseURL, endpoints, authToken);
        this.networkMonitorRef = React.createRef();

    }

    componentDidMount() {
        this.refresh();
        this.startAutoRefresh();
    }

    componentWillUnmount() {
        this.stopAutoRefresh();
    }

    refresh() {
        this.getAuthStatus();
        this.getUserList();
        this.getNetworkList();
        this.getVPNStatus();
        this.getStatisticList();
        this.getSystemStatus();
        this.getNetworkInterfaces();

        if (this.state.selectedInterface) {
            this.getInterfaceStats(this.state.selectedInterface);
        }
    }

    handleInterfaceChange = (interfaceName) => {
        this.setState({selectedInterface: interfaceName}, () => {
            if (interfaceName) {
                this.getInterfaceStats(interfaceName);
            }
        });
    };


    getAuthStatus() {
        this.api.call(
            "authStatus",
            {},
            true,
            this.handleGetAuthStatusSuccess.bind(this),
            this.handleGetAuthStatusFailure.bind(this)
        );
    }

    getUserList() {
        this.api.call(
            "userList",
            {},
            true,
            this.handleGetUsersSuccess.bind(this),
            this.handleGetUsersFailure.bind(this)
        );
    }

    getNetworkList() {
        this.api.call(
            "networkList",
            {},
            true,
            this.handleGetNetworksSuccess.bind(this),
            this.handleGetNetworksFailure.bind(this)
        );
    }

    getVPNStatus() {
        this.api.call(
            "vpnStatus",
            {},
            true,
            this.handleGetVPNStatusSuccess.bind(this),
            this.handleGetVPNStatusFailure.bind(this)
        );
    }

    getStatisticList() {
        this.api.call(
            "statisticList",
            {},
            true,
            this.handleGetStatisticListSuccess.bind(this),
            this.handleGetStatisticListFailure.bind(this)
        );
    }

    getSystemStatus() {
        this.api.call(
            "systemStatus",
            {},
            true,
            this.handleGetSystemStatusSuccess.bind(this),
            this.handleGetSystemStatusFailure.bind(this)
        );
    }

    getNetworkInterfaces() {
        this.api.call(
            "networkInterfaces",
            {},
            true,
            this.handleGetNetworkInterfacesSuccess.bind(this),
            this.handleGetNetworkInterfacesFailure.bind(this)
        );
    }

    getInterfaceStats(interfaceName) {
        if (!interfaceName) return;

        this.api.call(
            "interfaceStats",
            {name: interfaceName},
            true,
            this.handleGetInterfaceStatsSuccess.bind(this),
            this.handleGetInterfaceStatsFailure.bind(this)
        );
    }

    handleTabChange(i, value, tab, ev) {
        this.refresh();
    }

    handleGetUsersSuccess(res) {
        this.setState({users: res.data.users});
    }

    handleGetUsersFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        this.setState({users: []});
    }

    handleGetNetworksSuccess(res) {
        this.setState({networks: res.data.networks});
    }

    handleGetNetworksFailure(error) {
        console.log(error);
        this.setState({networks: []});
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleGetVPNStatusSuccess(res) {
        this.setState({vpn: res.data});
    }

    handleGetVPNStatusFailure(error) {
        console.log(error);
        this.setState({vpn: {}});
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleGetStatisticListSuccess(res) {
        this.setState({statistics: res.data.statistic});
    }

    handleGetStatisticListFailure(error) {
        console.log(error);
        this.setState({statistics: []});
        if (error.response && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleGetAuthStatusSuccess(res) {
        this.setState({self: res.data.user});
    }

    handleGetAuthStatusFailure(error) {
        console.log(error);
        this.setState({self: {}});
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleAuthFailure(error) {
        console.log("auth failure", error);
        ClearAuthToken();
        this.setState({logout: true});
    }

    handleGetSystemStatusFailure(error) {
        console.error("Failed to get system status:", error);
        if (error.response && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleGetNetworkInterfacesFailure(error) {
        console.error("Failed to get network interfaces:", error);
        if (error.response && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleGetInterfaceStatsFailure(error) {
        console.error("Failed to get interface stats:", error);
        if (error.response && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleCreateNewUser(e) {
        console.log("create new user");
        this.setState({modal: CREATINGNEWUSER});
    }

    handleDefineNewNetwork(e) {
        this.setState({modal: DEFININGNEWNETWORK});
    }

    handleUpdateUser(username, e) {
        for (let i in this.state.users) {
            if (this.state.users[i].username === username) {
                this.setState({modal: EDITINGUSER, editedUser: this.state.users[i]});
                console.log(i);
                console.log("updating user:", this.state.users[i].username);
                return;
            }
        }
    }

    handleCloseModal() {
        this.setState({modal: ""});
    }

    handleNewUserSave(user) {
        let userObj = {
            username: user.username,
            password: user.password,
            no_gw: user.pushGW,
            host_id: 0,
            is_admin: user.isAdmin
        };
        userObj.no_gw = !user.pushGW;
        userObj.admin_pref = user.isAdmin ? "ADMIN" : "NOADMIN";
        userObj.host_id =
            user.ipAllocationMethod === "static" ? dot2num(user.staticIP) : 0;
        userObj.static_pref =
            user.ipAllocationMethod === "static" ? "STATIC" : "NOSTATIC";

        console.log("creating new user:", user.username);
        this.api.call(
            "userCreate",
            userObj,
            true,
            this.handleCreateUserSuccess.bind(this),
            this.handleCreateUserFailure.bind(this)
        );
        this.setState({modal: ""});
    }

    handleCreateUserSuccess(res) {
        this.refresh();
    }

    handleCreateUserFailure(error) {
        console.log(error);
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleUpdateUserSave(user) {
        let updatedUser = {
            password: "",
            username: user.username,
            gwpref: "NOPREF",
            admin_pref: "NOPREFADMIN",
            static_pref: "NOPREFSTATIC",
            hostid: 0
        };

        if (user.password !== "") {
            updatedUser.password = user.password;
        }

        updatedUser.gwpref = user.pushGW ? "GW" : "NOGW";
        updatedUser.admin_pref = user.isAdmin ? "ADMIN" : "NOADMIN";
        updatedUser.host_id =
            user.ipAllocationMethod === "static" ? dot2num(user.staticIP) : 0;
        updatedUser.static_pref =
            user.ipAllocationMethod === "static" ? "STATIC" : "NOSTATIC";
        console.log("updating user:", updatedUser.username);
        this.api.call(
            "userUpdate",
            updatedUser,
            true,
            this.handleUpdateUserSuccess.bind(this),
            this.handleUpdateUserFailure.bind(this)
        );

        this.setState({modal: ""});
    }

    handleUpdateUserSuccess(res) {
        this.refresh();
    }

    handleUpdateUserFailure(error) {
        console.log(error);
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleRemoveUser(username) {
        if (username === this.state.self.username) {
            return;
        }
        this.api.call(
            "userDelete",
            {username: username},
            true,
            this.handleRemoveUserSuccess.bind(this),
            this.handleRemoveUserFailure.bind(this)
        );
    }

    handleRemoveUserSuccess(res) {
        this.refresh();
    }

    handleRemoveUserFailure(error) {
        console.log(error);
        if (error.response.status === 401) {
            this.handleAuthFailure(error);
        }
    }

    handleDownloadProfileClick(username, e) {
        this.setState({genConfigUsername: username});
        this.api.call(
            "genConfig",
            {username: username},
            true,
            this.handleDownloadProfileSuccess.bind(this),
            this.handleDownloadProfileFailure.bind(this)
        );
    }

    handleDownloadProfileSuccess(res) {
        let blob = res.data.client_config;
        saveData(blob, this.state.genConfigUsername + ".ovpn");
    }

    handleDownloadProfileFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        console.log(error);
    }

    handleDefineNetworkSave(network) {
        this.api.call(
            "netDefine",
            network,
            true,
            this.handleDefineNetworkSuccess.bind(this),
            this.handleDefineNetworkFailure.bind(this)
        );
        this.setState({modal: ""});
    }

    handleDefineNetworkSuccess(res) {
        this.refresh();
    }

    handleDefineNetworkFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        console.log(error);
    }

    handleUndefineNetwork(name) {
        this.api.call(
            "netUndefine",
            {name: name},
            true,
            this.handleUndefineNetworkSuccess.bind(this),
            this.handleUndefineNetworkFailure.bind(this)
        );
    }

    handleUndefineNetworkSuccess(res) {
        this.refresh();
    }

    handleUndefineNetworkFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        console.log(error);
    }

    handleAssociateUser(networkName) {
        let assocUsers = [];
        let network;
        for (let i in this.state.networks) {
            if (this.state.networks[i].name === networkName) {
                network = this.state.networks[i];
                break;
            }
        }
        for (let i in this.state.users) {
            let found = false;
            for (let j in network.associated_usernames) {
                if (this.state.users[i].username === network.associated_usernames[j]) {
                    found = true;
                }
            }
            if (!found) {
                assocUsers.push(this.state.users[i].username);
            }
        }
        this.setState({
            modal: ASSOCIATINGUSER,
            assocNetworkName: networkName,
            possibleAssocUsers: assocUsers
        });
    }

    handleDissociateUser(networkName) {
        let dissocUsers = [];
        let network;
        for (let i in this.state.networks) {
            if (this.state.networks[i].name === networkName) {
                network = this.state.networks[i];
                break;
            }
        }
        for (let i in this.state.users) {
            let found = false;
            for (let j in network.associated_usernames) {
                if (this.state.users[i].username === network.associated_usernames[j]) {
                    found = true;
                }
            }
            if (found) {
                dissocUsers.push(this.state.users[i].username);
            }
        }
        this.setState({
            modal: DISSOCIATINGUSER,
            dissocNetworkName: networkName,
            possibleDissocUsers: dissocUsers
        });
    }

    handleAssociateUserSave(username) {
        this.api.call(
            "netAssociate",
            {name: this.state.assocNetworkName, username: username},
            true,
            this.handleAssociateUserSuccess.bind(this),
            this.handleAssociateUserFailure.bind(this)
        );
        this.setState({modal: ""});
    }

    handleDissociateUserSave(username) {
        this.api.call(
            "netDissociate",
            {name: this.state.dissocNetworkName, username: username},
            true,
            this.handleDissociateUserSuccess.bind(this),
            this.handleDissociateUserFailure.bind(this)
        );
        this.setState({modal: ""});
    }

    handleAssociateUserSuccess(res) {
        this.refresh();
    }

    handleAssociateUserFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        console.log(error);
    }

    handleDissociateUserSuccess(res) {
        this.refresh();
    }

    handleDissociateUserFailure(error) {
        if ("response" in error && error.response.status === 401) {
            this.handleAuthFailure(error);
        }
        console.log(error);
    }

    handleRestartVPNServer() {
        this.api.call(
            "vpnRestart",
            {},
            true,
            function () {
                this.refresh();
            }.bind(this),
            function () {
                if ("response" in error && error.response.status === 401) {
                    this.handleAuthFailure(error);
                }
                console.log(error);
            }.bind(this)
        );
    }

    handleLogout() {
        ClearAuthToken();
        this.setState({logout: true});
    }

    startAutoRefresh() {
        this.stopAutoRefresh();

        if (this.state.autoRefreshEnabled) {
            this.refreshInterval = setInterval(() => {
                this.getUserList();
                this.getStatisticList();
                this.getSystemStatus();

                if (this.state.selectedInterface) {
                    this.getInterfaceStats(this.state.selectedInterface);
                }
            }, this.state.refreshInterval);
        }
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    toggleAutoRefresh() {
        this.setState(
            prevState => ({autoRefreshEnabled: !prevState.autoRefreshEnabled}),
            () => {
                if (this.state.autoRefreshEnabled) {
                    this.startAutoRefresh();
                } else {
                    this.stopAutoRefresh();
                }
            }
        );
    }

    handleRefreshIntervalChange(e) {
        const newInterval = parseInt(e.target.value) * 1000;
        this.setState({refreshInterval: newInterval}, () => {
            if (this.state.autoRefreshEnabled) {
                this.startAutoRefresh();
            }
        });
    }

    handleGetSystemStatusSuccess(res) {
        this.setState({systemStatus: res.data});
    }

    handleGetNetworkInterfacesSuccess(res) {
        this.setState({networkInterfaces: res.data.interfaces});
    }

    handleGetInterfaceStatsSuccess(res) {
        this.setState({interfaceStats: res.data});
        if (this.networkMonitorRef.current) {
            this.networkMonitorRef.current.updateStats(res.data.stats);
        }
    }

    render() {
        if (this.state.logout) {
            return <Redirect to="/login"/>;
        }

        let users = [];
        for (let i = 0; i < this.state.users.length; i++) {
            let isStatic = "";
            if (this.state.users[i].host_id !== 0) {
                isStatic = (
                    <small>
            <span
                className="glyphicon glyphicon glyphicon-pushpin"
                data-toggle="tooltip"
                title="Statically Allocated IP"
            ></span>
                    </small>
                );
            }

            let isAdmin;
            if (this.state.users[i].is_admin) {
                isAdmin = (
                    <small>
            <span
                className="glyphicon glyphicon-asterisk"
                data-toggle="tooltip"
                title="Admin"
            ></span>
                    </small>
                );
            }

            let noGW = (
                <span
                    className="glyphicon glyphicon-remove"
                    data-toggle="tooltip"
                    title="False"
                ></span>
            );
            if (!this.state.users[i].no_gw) {
                noGW = (
                    <span
                        className="glyphicon glyphicon-ok"
                        data-toggle="tooltip"
                        title="True"
                    ></span>
                );
            }

            let isOnline = (
                <span
                    className="text-muted"
                    style={{"fontSize": "2em", "verticalAlign": "middle"}}
                    data-toggle="tooltip"
                    title="Offline"
                >
          ◦
        </span>
            );
            if (this.state.users[i].is_connected) {
                let onlineSince =
                    "Online, since " +
                    moment(this.state.users[i].connected_since).fromNow() +
                    ".";
                console.log(onlineSince);
                isOnline = (
                    <span
                        className="text-success"
                        style={{"fontSize": "2em", "verticalAlign": "middle"}}
                        data-toggle="tooltip"
                        title={onlineSince}
                    >
            •
          </span>
                );
            }

            let certExpiry = (
                <span
                    className="glyphicon glyphicon-remove"
                    data-toggle="tooltip"
                    title="Expired"
                ></span>
            );
            if (moment(this.state.users[i].expires_at).isAfter(moment.now())) {
                let expiresIn =
                    "expires " + moment(this.state.users[i].expires_at).fromNow();
                certExpiry = (
                    <span data-toggle="tooltip" title={this.state.users[i].expires_at}>
            {expiresIn}
          </span>
                );
            }

            let createdAt = (
                <span data-toggle="tooltip" title={this.state.users[i].created_at}>
          {moment(this.state.users[i].created_at).fromNow()}
        </span>
            );

            let statistics;
            const user = this.state.users[i];

            if (user?.is_connected) {
                const formatTraffic = (bytes) => {
                    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
                    let traffic = parseFloat(bytes) || 0;
                    let unitIndex = 0;

                    while (traffic >= 1024 && unitIndex < units.length - 1) {
                        traffic /= 1024;
                        unitIndex++;
                    }

                    return {
                        value: traffic.toFixed(unitIndex === 0 ? 0 : 1),
                        unit: units[unitIndex],
                        rawBytes: parseFloat(bytes) || 0
                    };
                };

                const sent = formatTraffic(user.bytes_sent);
                const received = formatTraffic(user.bytes_received);
                const connectedSince = user.connected_since ? moment(user.connected_since).fromNow() : 'unknown';

                const txSpeed = formatTraffic(user.tx || 0);
                const rxSpeed = formatTraffic(user.rx || 0);

                statistics = (
                    <span
                        title={`Connected since: ${connectedSince}\nUpload: ${sent.value} ${sent.unit}\nDownload: ${received.value} ${received.unit}`}>
            🔗 📤{sent.value}{sent.unit} / 📥{received.value}{received.unit}
                        <br/>
            <span title="average for 5 sec">
                ↑{txSpeed.value}{txSpeed.unit} / ↓{rxSpeed.value}{rxSpeed.unit}
            </span>
        </span>
                );
            } else {
                statistics = <span title="Disconnected">❌</span>;
            }

            users.push(
                <tr key={"user" + i}>
                    <td>{i + 1}</td>
                    <td>
                        {isOnline} {this.state.users[i].username} {isAdmin}
                    </td>
                    <td>
                        {this.state.users[i].ip_net} {isStatic}
                    </td>
                    <td>{createdAt}</td>
                    <td>{certExpiry}</td>
                    <td className="mui--align-middle">{noGW}</td>
                    <td>{statistics}</td>
                    <td>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-floppy-save"
                  data-toggle="tooltip"
                  title="Download VPN Profile"
                  onClick={this.handleDownloadProfileClick.bind(
                      this,
                      this.state.users[i].username
                  )}
              ></span>
                        </a>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-edit"
                  data-toggle="tooltip"
                  title="Update User"
                  onClick={this.handleUpdateUser.bind(
                      this,
                      this.state.users[i].username
                  )}
              ></span>
                        </a>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-remove"
                  data-toggle="tooltip"
                  title="Delete User"
                  onClick={this.handleRemoveUser.bind(
                      this,
                      this.state.users[i].username
                  )}
              ></span>
                        </a>
                    </td>
                </tr>
            );
        }

        let networks = [];
        for (let i = 0; i < this.state.networks.length; i++) {
            let via;
            if (this.state.networks[i].type === "ROUTE") {
                via = "via vpn-server";

                if (this.state.networks[i].via && this.state.networks[i].via !== "") {
                    via = "via " + this.state.networks[i].via;
                }
            }
            networks.push(
                <tr key={"network" + i}>
                    <td>{i + 1}</td>
                    <td>{this.state.networks[i].name}</td>
                    <td>
                        {this.state.networks[i].cidr} {via}
                    </td>
                    <td>{this.state.networks[i].type}</td>
                    <td>{this.state.networks[i].created_at}</td>
                    <td>{this.state.networks[i].associated_usernames.join(", ")}</td>
                    <td>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-plus-sign"
                  data-toggle="tooltip"
                  onClick={this.handleAssociateUser.bind(
                      this,
                      this.state.networks[i].name
                  )}
                  title="Associate User"
              ></span>
                        </a>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-minus-sign"
                  data-toggle="tooltip"
                  onClick={this.handleDissociateUser.bind(
                      this,
                      this.state.networks[i].name
                  )}
                  title="Dissociate User"
              ></span>
                        </a>
                        <a style={{"paddingLeft": "5px"}}>
              <span
                  className="glyphicon glyphicon-remove"
                  data-toggle="tooltip"
                  onClick={this.handleUndefineNetwork.bind(
                      this,
                      this.state.networks[i].name
                  )}
                  title="Undefine Network"
              ></span>
                        </a>
                    </td>
                </tr>
            );
        }

        let statisticsRows = [];
        const formatTraffic = (bytes) => {
            const units = ['B', 'KB', 'MB', 'GB', 'TB'];
            let traffic = parseFloat(bytes) || 0;
            let unitIndex = 0;

            while (traffic >= 1024 && unitIndex < units.length - 1) {
                traffic /= 1024;
                unitIndex++;
            }

            return {
                value: traffic.toFixed(unitIndex === 0 ? 0 : 1),
                unit: units[unitIndex],
                rawBytes: parseFloat(bytes) || 0
            };
        };

        for (let i = 0; i < this.state.statistics.length; i++) {
            const stat = this.state.statistics[i];

            const totalReceived = formatTraffic(stat.total_bytes_received);
            const totalSent = formatTraffic(stat.total_bytes_sent);
            const totalTraffic = formatTraffic(stat.total_bytes);

            const avgDuration = parseFloat(stat.avg_connection_duration) || 0;
            let durationText = "N/A";

            if (avgDuration > 0) {
                if (avgDuration < 60) {
                    durationText = `${avgDuration.toFixed(0)} sec`;
                } else if (avgDuration < 3600) {
                    durationText = `${(avgDuration / 60).toFixed(1)} min`;
                } else {
                    durationText = `${(avgDuration / 3600).toFixed(1)} hours`;
                }
            }

            statisticsRows.push(
                <tr key={"statistic" + i}>
                    <td>{i + 1}</td>
                    <td>{stat.username}</td>
                    <td>{stat.connection_count}</td>
                    <td>
                        <span title={`${totalReceived.rawBytes} bytes`}>
                            📥 {totalReceived.value} {totalReceived.unit}
                        </span>
                    </td>
                    <td>
                        <span title={`${totalSent.rawBytes} bytes`}>
                            📤 {totalSent.value} {totalSent.unit}
                        </span>
                    </td>
                    <td>
                        <span title={`${totalTraffic.rawBytes} bytes`}>
                            🔄 {totalTraffic.value} {totalTraffic.unit}
                        </span>
                    </td>
                    <td>{durationText}</td>
                </tr>
            );
        }

        return (
            <>
                {}
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    padding: '10px',
                    background: '#fff',
                    borderBottom: '1px solid #e0e0e0'
                }}>
                    <div style={{width: '48%'}}>
                        <NetworkMonitor
                            ref={this.networkMonitorRef}
                            interfaces={this.state.networkInterfaces}
                            onInterfaceChange={this.handleInterfaceChange}
                        />
                    </div>
                    <div style={{width: '48%'}}>
                        <SystemMonitor systemStatus={this.state.systemStatus}/>
                    </div>
                </div>

                <Container style={{"paddingTop": "5%"}}>
                    <Panel>
                        <Button
                            className="mui--pull-right"
                            color="primary"
                            onClick={this.handleLogout.bind(this)}
                        >
                            Logout
                        </Button>
                        <Container>
                            <Modal
                                isOpen={this.state.modal === CREATINGNEWUSER}
                                contentLabel="Modal"
                                style={modalStyle}
                            >
                                <UserEdit
                                    title="Create New User"
                                    onCancel={this.handleCloseModal.bind(this)}
                                    onSave={this.handleNewUserSave.bind(this)}
                                    isUsernameDisabled={false}
                                />
                            </Modal>
                            <Modal
                                isOpen={this.state.modal === EDITINGUSER}
                                contentLabel="Modal"
                                style={modalStyle}
                            >
                                <UserEdit
                                    title="Update User"
                                    onCancel={this.handleCloseModal.bind(this)}
                                    onSave={this.handleUpdateUserSave.bind(this)}
                                    isUsernameDisabled={true}
                                    username={this.state.editedUser.username}
                                    isAdmin={this.state.editedUser.is_admin}
                                    pushGW={!this.state.editedUser.no_gw}
                                    ipAllocationMethod={
                                        this.state.editedUser.host_id === 0 ? "dynamic" : "static"
                                    }
                                    staticIP={
                                        this.state.editedUser.host_id === 0
                                            ? ""
                                            : num2dot(this.state.editedUser.host_id)
                                    }
                                />
                            </Modal>
                            <Modal
                                isOpen={this.state.modal === DEFININGNEWNETWORK}
                                contentLabel="Modal"
                                style={modalStyle}
                            >
                                <NetworkEdit
                                    title="New Network"
                                    onCancel={this.handleCloseModal.bind(this)}
                                    onSave={this.handleDefineNetworkSave.bind(this)}
                                />
                            </Modal>
                            <Modal
                                isOpen={this.state.modal === ASSOCIATINGUSER}
                                contentLabel="Modal"
                                style={modalStyle}
                            >
                                <UserPicker
                                    title="Associate User"
                                    onCancel={this.handleCloseModal.bind(this)}
                                    onSave={this.handleAssociateUserSave.bind(this)}
                                    userNames={this.state.possibleAssocUsers}
                                />
                            </Modal>
                            <Modal
                                isOpen={this.state.modal === DISSOCIATINGUSER}
                                contentLabel="Modal"
                                style={modalStyle}
                            >
                                <UserPicker
                                    title="Dissociate User"
                                    onCancel={this.handleCloseModal.bind(this)}
                                    onSave={this.handleDissociateUserSave.bind(this)}
                                    userNames={this.state.possibleDissocUsers}
                                />
                            </Modal>

                            <div>
                                <Tabs
                                    onChange={this.handleTabChange.bind(this)}
                                    defaultSelectedIndex={0}
                                >
                                    <Tab value="users" label="Users">
                                        <div className="mui--clearfix">
                                            <Button
                                                className="mui--pull-right"
                                                color="primary"
                                                onClick={this.handleCreateNewUser.bind(this)}
                                                style={{marginLeft: '10px'}}
                                            >
                                                + Create User
                                            </Button>

                                            <div className="mui--pull-right"
                                                 style={{display: 'flex', alignItems: 'center', gap: '10px'}}>
                                                <select
                                                    value={this.state.refreshInterval / 1000}
                                                    onChange={this.handleRefreshIntervalChange.bind(this)}
                                                    style={{
                                                        padding: '5px',
                                                        fontSize: '12px',
                                                        height: '30px'
                                                    }}
                                                >
                                                    <option value="1">1 sec</option>
                                                    <option value="5">5 sec</option>
                                                    <option value="10">10 sec</option>
                                                    <option value="30">30 sec</option>
                                                    <option value="60">60 sec</option>
                                                </select>

                                                <span
                                                    onClick={this.toggleAutoRefresh.bind(this)}
                                                    style={{
                                                        cursor: 'pointer',
                                                        fontSize: '18px',
                                                        opacity: this.state.autoRefreshEnabled ? 1 : 0.5,
                                                        textDecoration: this.state.autoRefreshEnabled ? 'none' : 'line-through'
                                                    }}
                                                    title={this.state.autoRefreshEnabled ? 'Auto-refresh enabled' : 'Auto-refresh disabled'}
                                                >
                        🔄
                      </span>
                                            </div>
                                        </div>

                                        {this.state.autoRefreshEnabled && (
                                            <div style={{marginBottom: '10px', fontSize: '12px', color: '#666'}}>
                                                Auto-refresh every {this.state.refreshInterval / 1000} seconds
                                            </div>
                                        )}

                                        <table className="mui-table mui-table--bordered mui--text-justify">
                                            <thead>
                                            <tr>
                                                <th></th>
                                                <th>USERNAME</th>
                                                <th>IP</th>
                                                <th>CREATED AT</th>
                                                <th>CERT</th>
                                                <th>PUSH GATEWAY</th>
                                                <th>STATISTICS</th>
                                                <th>ACTIONS</th>
                                            </tr>
                                            </thead>
                                            <tbody>{users}</tbody>
                                        </table>
                                    </Tab>
                                    <Tab value="networks" label="Networks">
                                        <Button
                                            className="mui--pull-right"
                                            color="primary"
                                            onClick={this.handleDefineNewNetwork.bind(this)}
                                        >
                                            + Define Net
                                        </Button>
                                        <table className="mui-table mui-table--bordered mui--text-justify">
                                            <thead>
                                            <tr>
                                                <th>#</th>
                                                <th>NAME</th>
                                                <th>CIDR</th>
                                                <th>TYPE</th>
                                                <th>CREATED AT</th>
                                                <th>ASSOC USERS</th>
                                                <th>ACTIONS</th>
                                            </tr>
                                            </thead>
                                            <tbody>{networks}</tbody>
                                        </table>
                                    </Tab>
                                    <Tab value="vpn" label="VPN">
                                        <Button
                                            className="mui--pull-right"
                                            color="primary"
                                            onClick={this.handleRestartVPNServer.bind(this)}
                                        >
                                            Restart VPN Server
                                        </Button>
                                        <table className="mui-table mui-table--bordered mui--text-justify">
                                            <thead>
                                            <tr>
                                                <th>KEY</th>
                                                <th>VALUE</th>
                                            </tr>
                                            </thead>
                                            <tbody>
                                            <tr>
                                                <td>Hostname</td>
                                                <td>{this.state.vpn.hostname}</td>
                                            </tr>
                                            <tr>
                                                <td>Proto</td>
                                                <td>{this.state.vpn.proto}</td>
                                            </tr>
                                            <tr>
                                                <td>Port</td>
                                                <td>{this.state.vpn.port}</td>
                                            </tr>
                                            <tr>
                                                <td>Network</td>
                                                {" "}
                                                <td>
                                                    {this.state.vpn.net} ({this.state.vpn.mask})
                                                </td>
                                            </tr>
                                            <tr>
                                                <td>DNS</td>
                                                <td>{this.state.vpn.dns}</td>
                                            </tr>
                                            <tr>
                                                <td>Created At</td>
                                                <td>{this.state.vpn.created_at}</td>
                                            </tr>
                                            <tr>
                                                <td>Expires At</td>
                                                <td>{this.state.vpn.expires_at}</td>
                                            </tr>
                                            <tr>
                                                <td>CA Expires At</td>
                                                <td>{this.state.vpn.ca_expires_at}</td>
                                            </tr>
                                            </tbody>
                                        </table>
                                    </Tab>
                                    <Tab value="statistics" label="Statistics">
                                        <div className="mui--clearfix">
                                            <div className="mui--pull-right"
                                                 style={{display: 'flex', alignItems: 'center', gap: '10px'}}>
                                                <select
                                                    value={this.state.refreshInterval / 1000}
                                                    onChange={this.handleRefreshIntervalChange.bind(this)}
                                                    style={{
                                                        padding: '5px',
                                                        fontSize: '12px',
                                                        height: '30px'
                                                    }}
                                                >
                                                    <option value="1">1 sec</option>
                                                    <option value="5">5 sec</option>
                                                    <option value="10">10 sec</option>
                                                    <option value="30">30 sec</option>
                                                    <option value="60">60 sec</option>
                                                </select>

                                                <span
                                                    onClick={this.toggleAutoRefresh.bind(this)}
                                                    style={{
                                                        cursor: 'pointer',
                                                        fontSize: '18px',
                                                        opacity: this.state.autoRefreshEnabled ? 1 : 0.5,
                                                        textDecoration: this.state.autoRefreshEnabled ? 'none' : 'line-through'
                                                    }}
                                                    title={this.state.autoRefreshEnabled ? 'Auto-refresh enabled' : 'Auto-refresh disabled'}
                                                >
                                                🔄
                                            </span>
                                            </div>
                                        </div>

                                        {this.state.autoRefreshEnabled && (
                                            <div style={{marginBottom: '10px', fontSize: '12px', color: '#666'}}>
                                                Auto-refresh every {this.state.refreshInterval / 1000} seconds
                                            </div>
                                        )}

                                        <table className="mui-table mui-table--bordered mui--text-justify">
                                            <thead>
                                            <tr>
                                                <th>#</th>
                                                <th>USERNAME</th>
                                                <th>CONNECTIONS</th>
                                                <th>TOTAL DOWNLOAD</th>
                                                <th>TOTAL UPLOAD</th>
                                                <th>TOTAL TRAFFIC</th>
                                                <th>AVG DURATION</th>
                                            </tr>
                                            </thead>
                                            <tbody>
                                            {statisticsRows.length > 0 ? statisticsRows : (
                                                <tr>
                                                    <td colSpan="7" className="mui--text-center">
                                                        No statistics available
                                                    </td>
                                                </tr>
                                            )}
                                            </tbody>
                                        </table>
                                    </Tab>
                                </Tabs>
                            </div>
                        </Container>
                    </Panel>
                </Container>
            </>
        );
    }
}
