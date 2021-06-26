import { castArray } from 'lodash'

function isDocument(element: any): element is Document {
    return element?.documentElement !== undefined
}

/**
 * An IntersectionObserver implementation that will immediately signal all
 * observed elements as intersecting.
 */
export class MockIntersectionObserver implements IntersectionObserver {
    public readonly root: Element | Document | null
    public readonly rootMargin: string
    public readonly thresholds: readonly number[]

    constructor(private callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
        this.root = options?.root ?? null
        this.rootMargin = options?.rootMargin ?? '0px 0px 0px 0px'
        this.thresholds = castArray(options?.threshold)
    }

    public observe(target: Element): void {
        this.callback(
            [
                {
                    target,
                    isIntersecting: true,
                    boundingClientRect: target.getBoundingClientRect(),
                    intersectionRect: target.getBoundingClientRect(),
                    rootBounds: isDocument(this.root)
                        ? document.documentElement.getBoundingClientRect()
                        : // Fallback on documentElement in case if root equals to null
                          (this.root ?? document.documentElement).getBoundingClientRect(),
                    intersectionRatio: 1,
                    time: 0,
                },
            ],
            this
        )
    }
    public takeRecords(): IntersectionObserverEntry[] {
        return []
    }
    public unobserve(): void {
        // noop
    }
    public disconnect(): void {
        // noop
    }
}
