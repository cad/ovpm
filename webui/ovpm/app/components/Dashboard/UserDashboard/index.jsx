import React from 'react';
import { Link } from 'react-router-dom'
import Panel from 'muicss/lib/react/panel';
import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';

import {GetAuthToken, ClearAuthToken} from '../../../utils/auth.js';
import {API} from '../../../utils/restClient.js';
import {baseURL, endpoints} from '../../../api.js';

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


export default class UserDashboard extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
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

    render() {
        return (
            <Container>
                <Panel>
                    <p>Welcome, <b>{this.props.username}</b> (<Link to="/logout">logout</Link>)!</p>
                    <div>
                        <Button color="primary" variant="raised" onClick={this.handleDownloadProfileClick.bind(this)}>Download VPN Profile</Button>
                    </div>
                </Panel>
            </Container>
        )
    }
}
