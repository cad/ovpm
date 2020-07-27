import React from "react";

import Button from "muicss/lib/react/button";
import Container from "muicss/lib/react/container";
import Option from "muicss/lib/react/option";
import Select from "muicss/lib/react/select";

export default class UserPicker extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      username: this.props.userNames.length > 0 ? this.props.userNames[0] : ""
    };
  }

  componentWillMount() {}

  handleUserChange(e) {
    this.setState({ username: e.target.value });
  }

  handleFormSubmit() {
    this.props.onSave(this.state.username);
    console.log(this.state.username);
  }

  handleFormCancel() {
    this.setState({ error: null });
    this.props.onCancel();
  }

  render() {
    let users = [];
    for (let i in this.props.userNames) {
      users.push(
        <Option
          key={i}
          value={this.props.userNames[i]}
          label={this.props.userNames[i]}
        />
      );
    }
    return (
      <Container>
        <h1>{this.props.title}</h1>

        <Select
          name="user"
          label="User"
          value={this.state.username}
          onChange={this.handleUserChange.bind(this)}
        >
          {users}
        </Select>
        <div className="mui--pull-right">
          <Button
            color="primary"
            onClick={this.handleFormSubmit.bind(this)}
            required={true}
          >
            Save
          </Button>
          <Button
            color="danger"
            onClick={this.handleFormCancel.bind(this)}
            required={true}
          >
            Cancel
          </Button>
        </div>
      </Container>
    );
  }
}
