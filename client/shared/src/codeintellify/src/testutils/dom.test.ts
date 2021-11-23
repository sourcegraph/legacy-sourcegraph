import { DOM } from './dom'

const { expect } = chai

describe('can create dom elements from generated code tables', () => {
    const dom = new DOM()
    after(dom.cleanup)

    it('can create the code view test cases and their helper function work', () => {
        for (const codeViewProps of dom.createCodeViews()) {
            const {
                codeView,
                getCodeElementFromTarget,
                getCodeElementFromLineNumber,
                getLineNumberFromCodeElement,
            } = codeViewProps

            for (let index = 1; index < 10; index++) {
                const cellFromLine = getCodeElementFromLineNumber(codeView, index)
                expect(cellFromLine).to.not.equal(null)
                const cellFromTarget = getCodeElementFromTarget(cellFromLine!)
                expect(cellFromTarget).to.equal(cellFromLine)
                const line = getLineNumberFromCodeElement(cellFromTarget!)
                expect(line).to.equal(index)
            }
        }
    })
})
