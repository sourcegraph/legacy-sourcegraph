import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { pluralize } from '../../../../shared/src/util/strings'
import { QueryFieldExamples } from './QueryBuilderInputRow'

interface Props {
    title: string
    markdown: string
    examples?: QueryFieldExamples[]
}

interface State {
    isOpen: boolean
}

export class InfoDropdown extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = { isOpen: false }
    }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))

    public render(): JSX.Element | null {
        return (
            <Dropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen} className="info-dropdown d-flex">
                <>
                    <DropdownToggle
                        tag="span"
                        caret={false}
                        className="pl-2 pr-0 btn btn-link d-flex align-items-center"
                    >
                        <HelpCircleOutlineIcon className="icon-inline small" />
                    </DropdownToggle>
                    <DropdownMenu right={true} className="pb-0 info-dropdown__item">
                        <DropdownItem header={true}>
                            <strong>{this.props.title}</strong>
                        </DropdownItem>
                        <DropdownItem divider={true} />
                        <div className="info-dropdown__content">
                            <small dangerouslySetInnerHTML={{ __html: renderMarkdown(this.props.markdown) }} />
                        </div>

                        {this.props.examples && (
                            <>
                                <DropdownItem divider={true} />
                                <DropdownItem header={true}>
                                    <strong>{pluralize('Example', this.props.examples.length)}</strong>
                                </DropdownItem>
                                <ul className="list-unstyled mb-2">
                                    {this.props.examples.map((ex: QueryFieldExamples) => (
                                        <div key={ex.value}>
                                            <div className="p-2">
                                                <span className="text-muted small">{ex.description}: </span>
                                                <code>{ex.value}</code>
                                            </div>
                                        </div>
                                    ))}
                                </ul>
                            </>
                        )}
                    </DropdownMenu>
                </>
            </Dropdown>
        )
    }
}
