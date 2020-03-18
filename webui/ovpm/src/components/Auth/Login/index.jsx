import React from 'react';
import ReactDOM from 'react-dom';
import Button from 'muicss/lib/react/button';
import Input from 'muicss/lib/react/input';
import Panel from 'muicss/lib/react/panel';
import Container from 'muicss/lib/react/container';

import { Redirect } from 'react-router'

import {IsAuthenticated, SetAuthToken, ClearAuthToken, SetItem, GetItem} from '../../../utils/auth.js';
import {API} from '../../../utils/restClient.js';
import {baseURL, endpoints} from '../../../api.js';

export default class Login extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            username: "",
            password: "",
            isAuthenticated: false,
            isAdmin: false,
            error: null,
        }
        this.api = new API(baseURL, endpoints)
    }

    componentWillMount() {
        let isAdmin = false
        if (GetItem("isAdmin")) {
            isAdmin = true
        }
        this.setState({
            isAuthenticated: IsAuthenticated(),
            isAdmin: isAdmin,
        })
        this.api.call("authStatus", {}, false, this.handleGetUserInfoSuccess.bind(this), this.handleGetUserInfoFailure.bind(this))
    }

    handleUsernameChange(e) {
        this.setState({username: e.target.value})
    }

    handlePasswordChange(e) {
        this.setState({password: e.target.value})
    }

    handleGetUserInfoSuccess(res) {
        if (res.data.user.username === "root") {
            SetAuthToken("root")
            this.setState({isAuthenticated: true})
            this.api.setAuthToken("root")
            SetItem("isAdmin", true)
            SetItem("username", "root")
        } else {
            SetItem("isAdmin", res.data.user.is_admin)
            SetItem("username", this.state.username)
        }
    }

    handleGetUserInfoFailure(error) {
        console.log(error)
    }

    handleAuthenticateSuccess(res) {
        SetAuthToken(res.data.token)
        this.setState({isAuthenticated: true})
        console.log("authenticated")
        this.api.setAuthToken(res.data.token)
        this.api.call("authStatus", {}, true, this.handleGetUserInfoSuccess.bind(this), this.handleGetUserInfoFailure.bind(this))
    }

    handleAuthenticateFailure(error) {
        ClearAuthToken()
        this.setState({isAuthenticated: false})
        console.log("authentication error", error)
        if (error.response.status >= 400) {
            this.setState({error: "Your credentials are incorrect."})
        }
    }

    handleFormSubmit(e) {
        this.setState({error: null})
        if (!this.state.username) {
            return
        }
        if (!this.state.password) {
            return
        }

        let data = {
            username: this.state.username,
            password: this.state.password
        }

        this.api.call("authenticate", data, false, this.handleAuthenticateSuccess.bind(this), this.handleAuthenticateFailure.bind(this))
        e.preventDefault()
    }
    render() {
        let error
        if (this.state.isAuthenticated) {
            return <Redirect to="/dashboard" />
        }

        if (this.state.error) {
            error = (
                <Panel className="mui--text-center" style={{"color": "#fff", "background-color": "#F44336", "margin-bottom":"0", "padding-bottom": "0"}}>
                    <b>Authentication Error</b>
                    <p>{this.state.error}</p>
                </Panel>)
        }
        return (
            <div style={{"maxWidth": "500px", "marginTop": "calc(50vh - 232px / 2)", "marginLeft": "calc(50% - 500px / 2)"}}>
                <Container>
                    {error}
                    <Panel>
                        <form onSubmit={this.handleFormSubmit.bind(this)}>
                            <Input label="Username" value={this.state.username} onChange={this.handleUsernameChange.bind(this)} floatingLabel={true} required={true} />
                            <Input label="Password" value={this.state.password} onChange={this.handlePasswordChange.bind(this)} floatingLabel={true} required={true} type="password" />
                            <Button type="submit" color="primary" required={true}>Login</Button>
                        </form>
                    </Panel>
                </Container>
            </div>
        )
    }
}
