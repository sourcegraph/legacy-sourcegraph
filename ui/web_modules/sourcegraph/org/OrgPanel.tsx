import * as React from "react";

import { Org, OrgMember } from "sourcegraph/api";
import { context } from "sourcegraph/app/context";
import { setLocationModalState } from "sourcegraph/components/Modal";
import { Spinner } from "sourcegraph/components/symbols";
import { whitespace } from "sourcegraph/components/utils/whitespace";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import { OrgInviteModal } from "sourcegraph/org/OrgInviteModal";
import { OrgMembersTable } from "sourcegraph/org/OrgMembersTable";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props {
	org: Org;
	members: OrgMember[] | null;
}

interface State {
	selectedMember: OrgMember | null;
	sentInvites: Array<String>;
}

export class OrgPanel extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			selectedMember: null,
			sentInvites: [],
		};
	}

	componentWillUpdate(nextProps: Props, nextState: State): void {
		if (this.props.members !== nextProps.members) {
			nextState.sentInvites = [];
		}
	}

	_invitedUser(member: OrgMember): void {
		if (member.Email != null && context.user != null && this.props.org.Login) {
			Dispatcher.Backends.dispatch(new OrgActions.SubmitOrgInvitation(member.Login || "", member.Email, this.props.org.Login, String(this.props.org.ID)));
			Events.OrgUser_Invited.logEvent({ org_name: this.props.org.Login, num_invites: 1 });
			this._updateSentInvites([member]);
		} else {
			Events.OrgManualInviteModal_Initiated.logEvent({ org_name: this.props.org.Login });
			setLocationModalState(this.context.router, "orgInvite", true);
			this.setState(Object.assign({}, this.state, {
				selectedMember: member,
			}));
		}
	}

	_onInviteUser(invites: Array<Object>): void {
		if (this.props.org && this.props.org.Login && context.user) {
			Events.OrgUser_Invited.logEvent({ org_name: this.props.org.Login, num_invites: invites.length });
			for (let i = 0; i < invites.length; i++) {
				let invite = invites[i];
				let member = invite["member"];
				Dispatcher.Backends.dispatch(new OrgActions.SubmitOrgInvitation(member["Login"] || "", invite["email"], this.props.org.Login, String(this.props.org.ID)));
			}

			setLocationModalState(this.context.router, "orgInvite", false);
			this._updateSentInvites(invites.map(invite => {
				return invite["member"];
			}));
		}
	}

	_updateSentInvites(members: OrgMember[]): void {
		let invites = this.state.sentInvites;
		let sentInvites = invites.concat(members.map(member => {
			return member.Login;
		}));
		this.setState(Object.assign({}, this.state, {
			sentInvites: sentInvites,
		}));
	}

	_orgMembersList(members: OrgMember[]): JSX.Element | null {
		if (members.length === 0) {
			return <div>
				<p>Looks like your organization is empty. Invite some of your users to join!</p>
			</div>;
		}

		return <OrgMembersTable sentInvites={this.state.sentInvites} inviteClicked={this._invitedUser.bind(this)} members={members} />;
	}

	render(): JSX.Element | null {
		let { members } = this.props;
		return <div style={{ paddingLeft: whitespace[5] }}>
			<OrgInviteModal onInvite={this._onInviteUser.bind(this)} member={this.state.selectedMember || null} org={this.props.org} />
			{members ?
				<div>{this._orgMembersList(members)}</div>
				: <div style={{ textAlign: "center", padding: whitespace[3] }}><Spinner /></div>
			}
		</div>;
	}
}
