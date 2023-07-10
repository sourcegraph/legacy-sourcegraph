import { Meta, Story } from '@storybook/react'

import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'

import { PostSignUpPage } from './PostSignUpPage'

const config: Meta = {
    title: 'web/src/auth/PostSignUpPage',
}

export default config

const mockUser = {
    __typename: 'User',
    id: '1',
    username: 'user',
    emails: [{ email: 'user@me.com', isPrimary: true, verified: false }],
    hasVerifiedEmail: false,
    completedPostSignup: false,
} as AuthenticatedUser

window.context.codyRequiresVerifiedEmail = true

export const UnverifiedEmail: Story = () => <WebStory>{() => <PostSignUpPage authenticatedUser={mockUser} />}</WebStory>

export const VerifiedEmail: Story = () => (
    <WebStory>{() => <PostSignUpPage authenticatedUser={{ ...mockUser, hasVerifiedEmail: true }} />}</WebStory>
)
