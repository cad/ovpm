import React from 'react';

import {GetAuthToken, ClearAuthToken} from '../../../utils/auth.js';
import {API} from '../../../utils/restClient.js';
import {baseURL, endpoints} from '../../../api.js';
import Panel from 'muicss/lib/react/panel';
import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';
import Tabs from 'muicss/lib/react/tabs';
import Tab from 'muicss/lib/react/tab';
import Input from 'muicss/lib/react/input';
import Modal from 'react-modal';
import UserEdit from './UserEdit';
import NetworkEdit from './NetworkEdit';
import UserPicker from './UserPicker';

const modalStyle = {
    content : {
        top                   : '50%',
        left                  : '50%',
        right                 : 'auto',
        bottom                : 'auto',
        marginRight           : '-50%',
        transform             : 'translate(-50%, -50%)'
    }
};
const CREATINGNEWUSER = "CREATINGNEWUSER"
const EDITINGUSER = "EDITINGUSER"
const DEFININGNEWNETWORK = "DEFININGNEWNETWORK"
const EDITINGNETWORK = "EDITINGNETWORK"
const ASSOCIATINGUSER = "ASSOCIATINGUSER"
const DISSOCIATINGUSER = "DISSOCIATINGUSER"

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
}());

function dot2num(dot)
{
    var d = dot.split('.');
    return ((((((+d[0])*256)+(+d[1]))*256)+(+d[2]))*256)+(+d[3]);
}

function num2dot(num)
{
    var d = num%256;
    for (var i = 3; i > 0; i--)
        {
            num = Math.floor(num/256);
            d = num%256 + '.' + d;
        }
    return d;
}

export default class AdminDashboard extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            users: [],
            networks: [],
            vpn: {},
            modal: "",
            self: {},
            editedUser: {},
            genConfigUsername: "",
            assocNetworkName: "",
            dissocNetworkName: "",
            possibleAssocUsers: [],
            possibleDissocUsers: [],
        }
        let authToken = GetAuthToken()
        this.api = new API(baseURL, endpoints, authToken)
    }

    componentWillMount() {
        this.refresh()
    }

    refresh() {
        this.getAuthStatus()
        this.getUserList()
        this.getNetworkList()
        this.getVPNStatus()
    }

    getAuthStatus() {
        this.api.call("authStatus", {}, true, this.handleGetAuthStatusSuccess.bind(this), this.handleGetAuthStatusFailure.bind(this))
    }

    getUserList() {
        this.api.call("userList", {}, true, this.handleGetUsersSuccess.bind(this), this.handleGetUsersFailure.bind(this))
    }

    getNetworkList() {
        this.api.call("networkList", {}, true, this.handleGetNetworksSuccess.bind(this), this.handleGetNetworksFailure.bind(this))
    }

    getVPNStatus() {
        this.api.call("vpnStatus", {}, true, this.handleGetVPNStatusSuccess.bind(this), this.handleGetVPNStatusFailure.bind(this))
    }

    handleTabChange (i, value, tab, ev) {
        this.refresh()
    }

    handleGetUsersSuccess(res) {
        this.setState({users: res.data.users})
    }

    handleGetUsersFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        this.setState({users: []})
    }

    handleGetNetworksSuccess(res) {
        this.setState({networks: res.data.networks})
    }

    handleGetNetworksFailure(error) {
        console.log(error)
        this.setState({networks: []})
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleGetVPNStatusSuccess(res) {
        this.setState({vpn: res.data})
    }

    handleGetVPNStatusFailure(error) {
        console.log(error)
        this.setState({vpn: {}})
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleGetAuthStatusSuccess(res) {
        this.setState({self: res.data.user})
    }

    handleGetAuthStatusFailure(error) {
        console.log(error)
        this.setState({self: {}})
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleAuthFailure(error) {
        console.log("auth failure", error)
        ClearAuthToken()
    }

    handleCreateNewUser(e) {
        this.setState({modal: CREATINGNEWUSER})
    }

    handleDefineNewNetwork(e) {
        this.setState({modal: DEFININGNEWNETWORK})
    }

    handleUpdateUser(username, e) {
        for (let i in this.state.users) {
            if (this.state.users[i].username === username) {
                this.setState({modal: EDITINGUSER, editedUser: this.state.users[i]})
                return
            }
        }
    }

    handleCloseModal() {
        this.setState({modal: ""})
    }

    handleNewUserSave(user) {
        console.log("HERE", user)
        let userObj = {
            username: user.username,
            password: user.password,
            no_gw: user.pushGW,
            host_id: 0, // handle this host_id problem
            is_admin: user.isAdmin,
        }
        userObj.gwpref = user.pushGW ? "GW" : "NOGW"
        userObj.admin_pref = user.isAdmin ? "ADMIN" : "NOADMIN"
        userObj.host_id = user.ipAllocationMethod === "static" ? dot2num(user.staticIP) : 0
        userObj.static_pref = user.ipAllocationMethod === "static" ? "STATIC" : "NOSTATIC"

        this.api.call("userCreate", userObj, true, this.handleCreateUserSuccess.bind(this), this.handleCreateUserFailure.bind(this))
        this.setState({modal: ""})
    }

    handleCreateUserSuccess(res) {
        this.refresh()
    }

    handleCreateUserFailure(error) {
        console.log(error)
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleUpdateUserSave(user) {
        let updatedUser = {
            password: "",
            username: user.username,
            gwpref: "NOPREF",
            admin_pref: "NOPREFADMIN",
            static_pref: "NOPREFSTATIC",
            hostid: 0,
        }

        if (user.password !== "") {
            updatedUser.password = user.password
        }

        updatedUser.gwpref = user.pushGW ? "GW" : "NOGW"
        updatedUser.admin_pref = user.isAdmin ? "ADMIN" : "NOADMIN"
        updatedUser.host_id = user.ipAllocationMethod === "static" ? dot2num(user.staticIP) : 0
        updatedUser.static_pref = user.ipAllocationMethod === "static" ? "STATIC" : "NOSTATIC"
        this.api.call("userUpdate", updatedUser, true, this.handleUpdateUserSuccess.bind(this), this.handleUpdateUserFailure.bind(this))

        this.setState({modal: ""})
    }

    handleUpdateUserSuccess(res) {
        this.refresh()
    }

    handleUpdateUserFailure(error) {
        console.log(error)
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }


    handleRemoveUser(username) {
        if (username === this.state.self.username) {
            // Don't remove yourself.
            return
        }
        this.api.call("userDelete", {username: username}, true, this.handleRemoveUserSuccess.bind(this), this.handleRemoveUserFailure.bind(this))
    }

    handleRemoveUserSuccess(res) {
        this.refresh()
    }

    handleRemoveUserFailure(error) {
        console.log(error)
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleDownloadProfileClick(username, e) {
        this.setState({genConfigUsername: username})
        this.api.call("genConfig", {username: username}, true, this.handleDownloadProfileSuccess.bind(this), this.handleDownloadProfileFailure.bind(this))
    }

    handleDownloadProfileSuccess(res) {
        let blob = res.data.client_config
        saveData(blob, this.state.genConfigUsername+".ovpn")
    }

    handleDownloadProfileFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }

    handleDefineNetworkSave(network) {
        console.log("NETWORK:", network)
        this.api.call("netDefine", network, true, this.handleDefineNetworkSuccess.bind(this), this.handleDefineNetworkFailure.bind(this))
        this.setState({modal: ""})
    }

    handleDefineNetworkSuccess(res) {
        this.refresh()
    }

    handleDefineNetworkFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }

    handleUndefineNetwork(name) {
        this.api.call("netUndefine", {name: name}, true, this.handleUndefineNetworkSuccess.bind(this), this.handleUndefineNetworkFailure.bind(this))
    }

    handleUndefineNetworkSuccess(res) {
        this.refresh()
    }

    handleUndefineNetworkFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }

    handleAssociateUser(networkName) {
        let assocUsers = []
        let network
        for(let i in this.state.networks) {
            if (this.state.networks[i].name === networkName) {
                network = this.state.networks[i]
                break
            }
        }
        for(let i in this.state.users) {
            let found = false
            for(let j in network.associated_usernames) {
                if(this.state.users[i].username === network.associated_usernames[j]) {
                    found = true
                }
            }
            if (!found) {
                assocUsers.push(this.state.users[i].username)
            }
        }
        this.setState({modal: ASSOCIATINGUSER, assocNetworkName: networkName, possibleAssocUsers: assocUsers})
    }

    handleDissociateUser(networkName) {
        let dissocUsers = []
        let network
        for(let i in this.state.networks) {
            if (this.state.networks[i].name === networkName) {
                network = this.state.networks[i]
                break
            }
        }
        for(let i in this.state.users) {
            let found = false
            for(let j in network.associated_usernames) {
                if(this.state.users[i].username === network.associated_usernames[j]) {
                    found = true
                }
            }
            if (found) {
                dissocUsers.push(this.state.users[i].username)
            }
        }
        this.setState({modal: DISSOCIATINGUSER, dissocNetworkName: networkName, possibleDissocUsers: dissocUsers})
    }

    handleAssociateUserSave(username) {
        //call
        //refresh
        console.log(username)
        this.api.call("netAssociate", {name: this.state.assocNetworkName, username : username}, true, this.handleAssociateUserSuccess.bind(this), this.handleAssociateUserFailure.bind(this))
        this.setState({modal: ""})
    }

    handleDissociateUserSave(username) {
        this.api.call("netDissociate", {name: this.state.dissocNetworkName, username : username}, true, this.handleDissociateUserSuccess.bind(this), this.handleDissociateUserFailure.bind(this))
        this.setState({modal: ""})
    }


    handleAssociateUserSuccess(res) {
        this.refresh()

    }

    handleAssociateUserFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }

    handleDissociateUserSuccess(res) {
        this.refresh()
    }

    handleDissociateUserFailure(error) {
        if ('response' in error && error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }





    render() {
        let users = []
        for (var i = 0; i < this.state.users.length; i++) {
            let isStatic = ""
            if (this.state.users[i].host_id != 0) {
                isStatic = (<small><span className="glyphicon glyphicon glyphicon-pushpin" data-toggle="tooltip" title="Statically Allocated IP"></span></small>)
            }

            let isAdmin
            if (this.state.users[i].is_admin) {
                isAdmin = (<small><span className="glyphicon glyphicon-asterisk" data-toggle="tooltip" title="Admin"></span></small>)
            }

            let noGW = (<span className="glyphicon glyphicon-remove" data-toggle="tooltip" title="False"></span>)
            if (!this.state.users[i].no_gw) {
                noGW = (<span className="glyphicon glyphicon-ok" data-toggle="tooltip" title="True"></span>)
            }

            users.push(
                <tr key={"user" + i}>
                    <td>{i+1}</td>
                    <td>{this.state.users[i].username} {isAdmin}</td>
                    <td>{this.state.users[i].ip_net} {isStatic}</td>
                    <td>{this.state.users[i].created_at}</td>
                    <td className="mui--align-middle">{noGW}</td>
                    <td>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-floppy-save" data-toggle="tooltip" title="Download VPN Profile" onClick={this.handleDownloadProfileClick.bind(this, this.state.users[i].username)}></span></a>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-edit" data-toggle="tooltip" title="Update User" onClick={this.handleUpdateUser.bind(this, this.state.users[i].username)}></span></a>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-remove" data-toggle="tooltip" title="Delete User" onClick={this.handleRemoveUser.bind(this, this.state.users[i].username)}></span></a>
                    </td>
                </tr>
            )
        }

        let networks = []
        for (var i = 0; i < this.state.networks.length; i++) {
            let via
            if (this.state.networks[i].type == "ROUTE") {
                via = "via vpn-server"

                if (this.state.networks[i].via && this.state.networks[i].via != "") {
                    via = "via " + this.state.networks[i].via
                }
            }
            networks.push(
                <tr key={"network" + i}>
                    <td>{i+1}</td>
                    <td>{this.state.networks[i].name}</td>
                    <td>{this.state.networks[i].cidr} {via}</td>
                    <td>{this.state.networks[i].type}</td>
                    <td>{this.state.networks[i].created_at}</td>
                    <td>{this.state.networks[i].associated_usernames.join(', ')}</td>
                    <td>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-plus-sign" data-toggle="tooltip" onClick={this.handleAssociateUser.bind(this, this.state.networks[i].name)} title="Associate User"></span></a>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-minus-sign" data-toggle="tooltip" onClick={this.handleDissociateUser.bind(this, this.state.networks[i].name)} title="Dissociate User"></span></a>
                        <a style={{"padding-left": "5px"}}><span className="glyphicon glyphicon-remove" data-toggle="tooltip" onClick={this.handleUndefineNetwork.bind(this, this.state.networks[i].name)} title="Undefine Network"></span></a>
                    </td>
                </tr>
            )
        }

        return (
            <Container>
                <Panel>
                    <Container>
                        <Modal isOpen={this.state.modal === CREATINGNEWUSER} contentLabel="Modal" style={modalStyle}>
                            <UserEdit title="Create New User" onCancel={this.handleCloseModal.bind(this)} onSave={this.handleNewUserSave.bind(this)} isUsernameDisabled={false}/>
                        </Modal>
                        <Modal isOpen={this.state.modal === EDITINGUSER} contentLabel="Modal" style={modalStyle}>
                            <UserEdit title="Update User" onCancel={this.handleCloseModal.bind(this)} onSave={this.handleUpdateUserSave.bind(this)} isUsernameDisabled={true} username={this.state.editedUser.username} isAdmin={this.state.editedUser.is_admin} pushGW={(!this.state.editedUser.no_gw)} ipAllocationMethod={this.state.editedUser.host_id == 0 ? "dynamic": "static"} staticIP={this.state.editedUser.host_id == 0 ? "": num2dot(this.state.editedUser.host_id)}/>
                        </Modal>
                        <Modal isOpen={this.state.modal === DEFININGNEWNETWORK} contentLabel="Modal" style={modalStyle}>
                            <NetworkEdit title="New Network" onCancel={this.handleCloseModal.bind(this)} onSave={this.handleDefineNetworkSave.bind(this)}/>
                        </Modal>
                        <Modal isOpen={this.state.modal === ASSOCIATINGUSER} contentLabel="Modal" style={modalStyle}>
                            <UserPicker title="Associate User" onCancel={this.handleCloseModal.bind(this)} onSave={this.handleAssociateUserSave.bind(this)} userNames={this.state.possibleAssocUsers} />
                        </Modal>
                        <Modal isOpen={this.state.modal === DISSOCIATINGUSER} contentLabel="Modal" style={modalStyle}>
                            <UserPicker title="Dissociate User" onCancel={this.handleCloseModal.bind(this)} onSave={this.handleDissociateUserSave.bind(this)} userNames={this.state.possibleDissocUsers} />
                        </Modal>



                        <div>
                            <Tabs onChange={this.handleTabChange.bind(this)} defaultSelectedIndex={0}>
                                <Tab value="users" label="Users">
                                    <Button className="mui--pull-right" color="primary" onClick={this.handleCreateNewUser.bind(this)}>+ Create User</Button>
                                    <table className="mui-table mui-table--bordered mui--text-justify">
                                        <thead>
                                            <tr>
                                                <th>#</th>
                                                <th>USERNAME</th>
                                                <th>IP</th>
                                                <th>CREATED AT</th>
                                                <th>PUSH GATEWAY</th>
                                                <th>ACTIONS</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {users}
                                        </tbody>
                                    </table>
                                </Tab>
                                <Tab value="networks" label="Networks">
                                    <Button className="mui--pull-right" color="primary" onClick={this.handleDefineNewNetwork.bind(this)}>+ Define Net</Button>
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
                                        <tbody>
                                            {networks}
                                        </tbody>
                                    </table>
                                </Tab>
                                <Tab value="vpn" label="VPN">
                                    <table className="mui-table mui-table--bordered mui--text-justify">
                                        <thead>
                                            <tr>
                                                <th>KEY</th>
                                                <th>VALUE</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <tr><td>Hostname</td>  <td>{this.state.vpn.hostname}</td></tr>
                                            <tr><td>Proto</td>     <td>{this.state.vpn.proto}</td></tr>
                                            <tr><td>Port</td>      <td>{this.state.vpn.port}</td></tr>
                                            <tr><td>Network</td>   <td>{this.state.vpn.net} ({this.state.vpn.mask})</td></tr>
                                            <tr><td>DNS</td>       <td>{this.state.vpn.dns}</td></tr>
                                            <tr><td>Created At</td><td>{this.state.vpn.created_at}</td></tr>
                                        </tbody>
                                    </table>
                                </Tab>
                            </Tabs>
                        </div>
                    </Container>
                </Panel>
            </Container>
        )
}
}
