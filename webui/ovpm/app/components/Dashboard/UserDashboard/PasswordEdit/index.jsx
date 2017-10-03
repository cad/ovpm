import React from 'react';

import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';
import Input from 'muicss/lib/react/input';


export default class UserEdit extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            password: "",
        }

    }

    componentWillMount() {
    }

    handlePasswordChange(e) {
        this.setState({password: e.target.value})
    }

    handleFormSubmit() {
        this.props.onSave(this.state.password)
    }

    handleFormCancel() {
        this.setState({error: null})
        this.props.onCancel()
    }

    render() {
        return (
            <Container>
                <h1>{this.props.title}</h1>

                <Input label="Password" value={this.state.password} onChange={this.handlePasswordChange.bind(this)} floatingLabel={true} required={true} type="password"/>
                <div className="mui--pull-right">
                    <Button color="primary" onClick={this.handleFormSubmit.bind(this)} required={true}>Save</Button>
                    <Button color="danger" onClick={this.handleFormCancel.bind(this)} required={true}>Cancel</Button>
                </div>
            </Container>
        )
    }
}
