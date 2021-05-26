import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { PageTitle } from '../../../components/PageTitle'

import combySample from './samples/comby.batch.yaml'
import helloWorldSample from './samples/empty.batch.yaml'
import goImportsSample from './samples/go-imports.batch.yaml'
import minimalSample from './samples/minimal.batch.yaml'

interface SampleTabHeaderProps {
    sample: Sample
    active: boolean
    setSelectedSample: (sample: Sample) => void
}

const SampleTabHeader: React.FunctionComponent<SampleTabHeaderProps> = ({ sample, active, setSelectedSample }) => {
    const onClick = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedSample(sample)
        },
        [setSelectedSample, sample]
    )
    return (
        <li className="nav-item">
            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
            <a href="" onClick={onClick} className={classNames('nav-link', active && 'active')} role="button">
                {sample.name}
            </a>
        </li>
    )
}

interface Sample {
    name: string
    file: string
}

const samples: Sample[] = [
    { name: 'hello-world.batch.yaml', file: helloWorldSample },
    { name: 'modify-with-comby.batch.yaml', file: combySample },
    { name: 'update-go-imports.batch.yaml', file: goImportsSample },
    { name: 'minimal.batch.yaml', file: minimalSample },
]

export interface CreateBatchChangePageProps {
    // Nothing for now.
}

export const CreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = () => {
    const [selectedSample, setSelectedSample] = useState<Sample>(samples[0])
    return (
        <>
            <PageTitle title="Create batch change" />
            <PageHeader
                path={[{ icon: BatchChangesIcon, text: 'Create batch change' }]}
                headingElement="h2"
                className="mb-3"
            />
            <Container className="mb-3">
                <h2>1. Write a batch spec YAML file</h2>
                <p className="mb-0">
                    The batch spec (
                    <a
                        href="https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </a>
                    ) describes what the batch change does. You'll provide it when previewing, creating, and updating
                    batch changes. We recommend committing it to source control.
                </p>
            </Container>
            <div className="row mb-3">
                <div className="col-4">
                    <h3>Examples</h3>
                    <ul className="nav nav-pills">
                        {samples.map(sample => (
                            <SampleTabHeader
                                key={sample.name}
                                sample={sample}
                                active={selectedSample.name === sample.name}
                                setSelectedSample={setSelectedSample}
                            />
                        ))}
                    </ul>
                </div>
                <div className="col-8">
                    <Container>
                        <CodeSnippet code={selectedSample.file} language="yaml" className="mb-4" />
                    </Container>
                </div>
            </div>
            <Container className="mb-3">
                <h2>2. Preview the batch change with Sourcegraph CLI</h2>
                <p>
                    Use the{' '}
                    <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        Sourcegraph CLI (src)
                    </a>{' '}
                    to preview the commits and changesets that your batch change will make:
                </p>
                <CodeSnippet code={`src batch preview -f ${selectedSample.name}`} language="bash" className="mb-3" />
                <p className="mb-0">
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the batch
                    change.
                </p>
            </Container>
            <p>
                Want more help? See{' '}
                <a href="/help/batch_changes" rel="noopener noreferrer" target="_blank">
                    Batch Changes documentation
                </a>
                .
            </p>
        </>
    )
}
