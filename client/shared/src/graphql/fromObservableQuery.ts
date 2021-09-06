import { ApolloQueryResult, ObservableQuery } from '@apollo/client'
import { Observable } from 'rxjs'

/**
 * Converts ObservableQuery returned by `client.watchQuery` to `rxjs` Observable.
 *
 * ```ts
 * const rxjsObservable = fromObservableQuery(client.watchQuery(query))
 * ```
 */
export function fromObservableQuery<T extends object>(
    observableQuery: ObservableQuery<T>
): Observable<ApolloQueryResult<T>> {
    return new Observable<ApolloQueryResult<T>>(subscriber => {
        const subscription = observableQuery.subscribe(subscriber)

        return function unsubscribe() {
            subscription.unsubscribe()
        }
    })
}

/**
 * Converts Promise<ObservableQuery> to `rxjs` Observable.
 *
 * ```ts
 * const rxjsObservable = fromObservableQuery(
 *   getGraphqlClient().then(client => client.watchQuery(query))
 * )
 * ```
 */
export function fromObservableQueryPromise<T extends object, V extends object>(
    observableQueryPromise: Promise<ObservableQuery<T, V>>
): Observable<ApolloQueryResult<T>> {
    return new Observable<ApolloQueryResult<T>>(subscriber => {
        const subscriptionPromise = observableQueryPromise.then(observableQuery =>
            observableQuery.subscribe(subscriber)
        )

        return function unsubscribe() {
            subscriber.unsubscribe()
            subscriptionPromise.then(subscription => subscription.unsubscribe()).catch(error => console.log(error))
        }
    })
}
