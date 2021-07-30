import { Meta, Story } from '@storybook/react'
import React from 'react'

import { Grid, GridProps } from './Grid'

export const GridExample: Story<GridProps> = ({ columnCount = 5, spacing }) => (
    <Grid columnCount={columnCount} spacing={spacing}>
        {/* Fill the grid with 20 items to showcase different setups */}
        {new Array(20).fill(0).map((_value, index) => (
            <div key={index}>Column</div>
        ))}
    </Grid>
)

const config: Meta = {
    title: 'wildcard/Grid',

    parameters: {
        component: Grid,
    },
    argTypes: {
        columnCount: {
            type: 'number',
        },
        spacing: {
            type: 'number',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default config
