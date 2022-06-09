import { useLocation } from 'react-router'

import helloWorldSample from '../batch-spec/edit/library/hello-world.batch.yaml'
import { insertQueryIntoLibraryItem, insertNameIntoLibraryItem } from '../batch-spec/yaml-util'

interface UseSearchQueryResult {
    renderTemplate?: (title: string) => string
}

const createRenderTemplate = (query: string): ((title: string) => string) => {
    const template = insertQueryIntoLibraryItem(helloWorldSample, query)

    return title => insertNameIntoLibraryItem(template, title)
}

/**
 * Custom hook for create page which creates a batch spec from a search query
 */
export const useSearchTemplate = (): UseSearchQueryResult => {
    const location = useLocation()
    const parameters = new URLSearchParams(location.search)

    const query = parameters.get('q')
    const patternType = parameters.get('patternType')

    if (query) {
        const searchQuery = `${query} ${patternType ? `patternType:${patternType}` : ''}`
        const renderTemplate = createRenderTemplate(searchQuery)
        return { renderTemplate }
    }

    return { renderTemplate: undefined }
}
