import { render } from '@testing-library/react'
import React from 'react'

import { Button } from './Button'
import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'

describe('Button', () => {
    it('renders a simple button correctly', () => {
        const { asFragment } = render(<Button>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it('supports rendering as different elements', () => {
        const { asFragment } = render(<Button as="a">I am a link</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it('renders a special tooltip sibling if the button is disabled and has a tooltip', () => {
        const { container } = render(
            <Button data-tooltip="I am the tooltip" disabled={true}>
                Disabled
            </Button>
        )
        expect(container.firstChild).toMatchInlineSnapshot(`
            <div
              class="container"
            >
              <div
                class="disabledTooltip"
                data-tooltip="I am the tooltip"
                tabindex="0"
              />
              <button
                class="btn"
                data-tooltip="I am the tooltip"
                disabled=""
                type="button"
              >
                Disabled
              </button>
            </div>
        `)
    })

    it.each(BUTTON_VARIANTS)("Renders variant '%s' correctly", variant => {
        const { asFragment } = render(<Button variant={variant}>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it.each(BUTTON_SIZES)("Renders size '%s' correctly", size => {
        const { asFragment } = render(<Button size={size}>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })
})
