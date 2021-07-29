import React from 'react'
import { BaseControlInput } from '../BaseControlInput'

export const RadioButton: typeof BaseControlInput = React.forwardRef((props, reference) => (
    <BaseControlInput {...props} type="radio" ref={reference} />
))
