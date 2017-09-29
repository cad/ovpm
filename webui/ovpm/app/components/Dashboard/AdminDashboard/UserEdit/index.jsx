import React from 'react';

import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';
import Input from 'muicss/lib/react/input';
import Option from 'muicss/lib/react/option';
import Select from 'muicss/lib/react/select';
import Checkbox from 'muicss/lib/react/checkbox';


export default class UserEdit extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            username: this.props.username ? this.props.username : "",
            password: "",
            staticIP: this.props.staticIP ? this.props.staticIP : "",
            ipAllocationMethod: this.props.ipAllocationMethod ? this.props.ipAllocationMethod : "dynamic",
            pushGW: this.props.pushGW ? this.props.pushGW : false,
            isAdmin: this.props.isAdmin ? this.props.isAdmin : false,
        }

    }

    componentWillMount() {
    }

    handleUsernameChange(e) {
        this.setState({username: e.target.value})
    }

    handlePasswordChange(e) {
        this.setState({password: e.target.value})
    }

    handleStaticIPChange(e) {
        this.setState({staticIP: e.target.value})
    }

    handleIPAllocationChange(e) {
        this.setState({ipAllocationMethod: e.target.value})
    }

    handlePushGWChange (e) {
        this.setState({pushGW: e.target.checked})
    }

    handleIsAdminChange (e) {
        this.setState({isAdmin: e.target.checked})
    }

    handleFormSubmit() {
        this.props.onSave(this.state)
    }

    handleFormCancel() {
        this.setState({error: null})
        this.props.onCancel()
    }

    render() {
        var staticIPInput
        if (this.state.ipAllocationMethod === "static") {
            staticIPInput = <Input label="Address" value={this.state.staticIP} onChange={this.handleStaticIPChange.bind(this)} floatingLabel={true} required={true} />
        }
        return (
            <Container>
                <h1>{this.props.title}</h1>

                <Input label="Username" value={this.state.username} onChange={this.handleUsernameChange.bind(this)} floatingLabel={true} required={true} disabled={this.props.isUsernameDisabled}/>
                <Input label="Password" value={this.state.password} onChange={this.handlePasswordChange.bind(this)} floatingLabel={true} required={true} type="password"/>
                <Select name="ip" label="IP Allocation" value={this.state.ipAllocationMethod} onChange={this.handleIPAllocationChange.bind(this)}>
                    <Option value="dynamic" label="Dynamic" />
                    <Option value="static" label="Static" />
                </Select>
                {staticIPInput}
                <Checkbox label="Push GW" checked={this.state.pushGW} onChange={this.handlePushGWChange.bind(this)}/>
                <Checkbox label="Make Admin" checked={this.state.isAdmin} onChange={this.handleIsAdminChange.bind(this)}/>
                <div className="mui--pull-right">
                    <Button color="primary" onClick={this.handleFormSubmit.bind(this)} required={true}>Save</Button>
                    <Button color="danger" onClick={this.handleFormCancel.bind(this)} required={true}>Cancel</Button>
                </div>
            </Container>
        )
    }
}
