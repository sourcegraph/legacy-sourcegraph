import React from "react";

class Component extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
	}

	componentWillMount() {
		this._updateState(Object.assign({}, this.state), this.props);
	}

	componentWillReceiveProps(nextProps) {
		this._updateState(Object.assign({}, this.state), nextProps);
	}

	shouldComponentUpdate(nextProps, nextState) {
		let keys = Object.keys(nextState);
		if (Object.keys(this.state).length !== keys.length) {
			return true;
		}
		for (let i = 0; i < keys.length; i++) {
			let k = keys[i];
			if (nextState[k] !== this.state[k]) {
				return true;
			}
		}
		return false;
	}

	setState(patch, callback) {
		this._updateState(Object.assign({}, this.state, patch), this.props, callback);
	}

	_updateState(newState, props, callback) {
		this._checkForUndefined(props, "Property");
		this.reconcileState(newState, props);
		this._checkForUndefined(newState, "State");
		if (this.onStateTransition) {
			this.onStateTransition(this.state, newState);
		}
		super.setState(newState, callback);
	}

	_checkForUndefined(obj, type) {
		if (process.env.NODE_ENV === "production") { return; }
		let keys = Object.keys(obj);
		for (let i = 0; i < keys.length; i++) {
			if (obj[keys[i]] === undefined) { // eslint-disable-line no-undefined
				throw new Error(`Invariant Violation: ${type} '${keys[i]}' of ${this.constructor.name} has value 'undefined'.`);
			}
		}
	}

	reconcileState(state, props) {
		// override
	}
}

export default Component;
