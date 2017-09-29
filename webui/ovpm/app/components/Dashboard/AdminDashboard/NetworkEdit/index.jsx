import React from 'react';

import Button from 'muicss/lib/react/button';
import Container from 'muicss/lib/react/container';
import Input from 'muicss/lib/react/input';
import Option from 'muicss/lib/react/option';
import Select from 'muicss/lib/react/select';
import Checkbox from 'muicss/lib/react/checkbox';


export default class NetworkEdit extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            name: this.props.name ? this.props.name : "",
            type: this.props.type ? this.props.type : "SERVERNET",
            cidr: this.props.cidr ? this.props.cidr : "",
            via: this.props.via ? this.props.via : "",
        }
    }

    componentWillMount() {
    }

    handleNameChange(e) {
        this.setState({name: e.target.value})
    }

    handleTypeChange(e) {
        this.setState({type: e.target.value})
    }

    handleCidrChange(e) {
        this.setState({cidr: e.target.value})
    }

    handleViaChange(e) {
        this.setState({via: e.target.value})
    }

    handleFormSubmit() {
        console.log(this.state.type)
        let network = {
            name: this.state.name,
            cidr: this.state.cidr,
            type: this.state.type,
            via: this.state.via,
        }

        this.props.onSave(network)
    }

    handleFormCancel() {
        this.setState({error: null})
        this.props.onCancel()
    }

    render() {
        var via
        if (this.state.type === "ROUTE") {
            via = <Input label="Via (Optional)" value={this.state.via} onChange={this.handleViaChange.bind(this)} floatingLabel={true} required={false} />
        }

        return (
            <Container>
                <h1>{this.props.title}</h1>

                <Input label="Name" value={this.state.name} onChange={this.handleNameChange.bind(this)} floatingLabel={true} required={true} disabled={this.props.isNameDisabled}/>
                <Input label="CIDR" value={this.state.cidr} onChange={this.handleCidrChange.bind(this)} floatingLabel={true} required={true} />
                <Select name="type" label="Type" value={this.state.type} onChange={this.handleTypeChange.bind(this)}>
                    <Option value="SERVERNET" label="SERVERNET" />
                    <Option value="ROUTE" label="ROUTE" />
                </Select>
                {via}

                <div className="mui--pull-right">
                    <Button color="primary" onClick={this.handleFormSubmit.bind(this)} required={true}>Save</Button>
                    <Button color="danger" onClick={this.handleFormCancel.bind(this)} required={true}>Cancel</Button>
                </div>
            </Container>
        )
    }
}
