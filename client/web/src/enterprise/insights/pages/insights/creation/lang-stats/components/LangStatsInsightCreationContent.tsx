import { FC, ReactNode } from 'react'

import { noop } from 'rxjs'

import {
    CreationUiLayout,
    CreationUIForm,
    CreationUIPreview,
    useField,
    FormChangeEvent,
    SubmissionErrors,
    useForm,
    insightTitleValidator,
    createRequiredValidator,
    insightRepositoriesValidator,
    insightRepositoriesAsyncValidator,
} from '../../../../../components'
import { LangStatsCreationFormFields } from '../types'

import {
    LangStatsInsightCreationForm,
    RenderPropertyInputs,
} from './lang-stats-insight-creation-form/LangStatsInsightCreationForm'
import { LangStatsInsightLivePreview } from './live-preview-chart/LangStatsInsightLivePreview'

export const thresholdFieldValidator = createRequiredValidator('Threshold is a required field for code insight.')

const INITIAL_VALUES: LangStatsCreationFormFields = {
    repository: '',
    title: '',
    threshold: 3,
    dashboardReferenceCount: 0,
}

export interface LangStatsInsightCreationContentProps {
    touched: boolean
    children: (input: RenderPropertyInputs) => ReactNode
    initialValues?: Partial<LangStatsCreationFormFields>
    className?: string

    onSubmit: (values: LangStatsCreationFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void

    /** Change handlers is called every time when user changed any field within the form. */
    onChange?: (event: FormChangeEvent<LangStatsCreationFormFields>) => void
}

export const LangStatsInsightCreationContent: FC<LangStatsInsightCreationContentProps> = props => {
    const { touched, initialValues = {}, className, children, onSubmit, onChange = noop } = props

    const { values, handleSubmit, formAPI } = useForm<LangStatsCreationFormFields>({
        initialValues: {
            ...INITIAL_VALUES,
            ...initialValues,
        },
        onSubmit,
        onChange,
        touched,
    })

    const repository = useField({
        name: 'repository',
        formApi: formAPI,
        validators: {
            sync: insightRepositoriesValidator,
            async: insightRepositoriesAsyncValidator,
        },
    })

    const title = useField({
        name: 'title',
        formApi: formAPI,
        validators: { sync: insightTitleValidator },
    })

    const threshold = useField({
        name: 'threshold',
        formApi: formAPI,
        validators: { sync: thresholdFieldValidator },
    })

    // If some fields that needed to run live preview  are invalid
    // we should disable live chart preview
    const allFieldsForPreviewAreValid = repository.meta.validState === 'VALID' && threshold.meta.validState === 'VALID'

    const handleFormReset = (): void => {
        // TODO [VK] Change useForm API in order to implement form.reset method.
        title.input.onChange('')
        repository.input.onChange('')
        // Focus first element of the form
        repository.input.ref.current?.focus()
        threshold.input.onChange(3)
    }

    const hasFilledValue = values.repository !== '' || values.title !== ''

    return (
        <CreationUiLayout data-testid="code-stats-insight-creation-page-content" className={className}>
            <CreationUIForm
                as={LangStatsInsightCreationForm}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                title={title}
                repository={repository}
                threshold={threshold}
                isFormClearActive={hasFilledValue}
                dashboardReferenceCount={initialValues.dashboardReferenceCount}
                onFormReset={handleFormReset}
            >
                {children}
            </CreationUIForm>

            <CreationUIPreview
                as={LangStatsInsightLivePreview}
                repository={repository.meta.value}
                threshold={threshold.meta.value}
                disabled={!allFieldsForPreviewAreValid}
            />
        </CreationUiLayout>
    )
}
