import React from 'react';
import { Link } from 'react-router-dom'
import Panel from 'muicss/lib/react/panel';
import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';

import {GetAuthToken, ClearAuthToken} from '../../../utils/auth.js';
import {API} from '../../../utils/restClient.js';
import {baseURL, endpoints} from '../../../api.js';

import Modal from 'react-modal';
import PasswordEdit from './PasswordEdit';

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


export default class UserDashboard extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            isChangePasswordModalOpen: false,
        }
        let authToken = GetAuthToken()
        this.api = new API(baseURL, endpoints, authToken)
    }

    componentWillMount() {
    }

    handleDownloadProfileClick(e) {
        this.props.api.call("genConfig", {username: this.props.username}, true, this.handleDownloadProfileSuccess.bind(this), this.handleDownloadProfileFailure.bind(this))
    }

    handleDownloadProfileSuccess(res) {
        let blob = res.data.client_config
        saveData(blob, this.props.username+".ovpn")
    }

    handleDownloadProfileFailure(error) {
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
        console.log(error)
    }

    handleAuthFailure(error) {
        console.log("auth failure", error)
        ClearAuthToken()
    }

    handleCloseModal() {
        this.setState({isChangePasswordModalOpen: false})
    }
    handleOpenChangePasswordModal() {
        this.setState({isChangePasswordModalOpen: true})
    }

    handleChangePasswordSave(password) {
        let userObj = {
            username: this.props.username,
            password: password,
        }
        this.api.call("userUpdate", userObj, true, this.handleChangePasswordSuccess.bind(this), this.handleChangePasswordFailure.bind(this))
        this.handleCloseChangePasswordModal()
    }

    handleChangePasswordSuccess(res) {
        console.log("password changed")
    }

    handleChangePasswordFailure(error) {
        console.log(error)
        if (error.response.status == 401) {
            this.handleAuthFailure(error)
        }
    }

    handleAuthFailure(error) {
        console.log("auth failure", error)
        ClearAuthToken()
    }

    handleCloseChangePasswordModal() {
        this.setState({isChangePasswordModalOpen: false})
    }

    render() {
        let passwordResetModal = (
            <Modal isOpen={this.state.isChangePasswordModalOpen} contentLabel="Modal" style={modalStyle}>
                <PasswordEdit title="Change Password" onCancel={this.handleCloseChangePasswordModal.bind(this)} onSave={this.handleChangePasswordSave.bind(this)} />
            </Modal>
        )

        return (
            <Container>
                {passwordResetModal}
                <Panel>
                    <p>Welcome, <b>{this.props.username}</b> (<Link to="/logout">logout</Link>)!</p>
                    <div>
                        <Button color="primary" variant="raised" onClick={this.handleOpenChangePasswordModal.bind(this)}>Change Password</Button>
                        <Button color="primary" variant="raised" onClick={this.handleDownloadProfileClick.bind(this)}>Download VPN Profile</Button>
                    </div>
                </Panel>
            </Container>
        )
    }
}
