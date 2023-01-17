const MAX_HEIGHT = 192

class Controller {
	constructor(vscode, model) {
		this.vscode = vscode
		this.model = model
		this.chatContainer = document.querySelector('.chat-container')
		this.recipesContainer = document.querySelector('.recipes-container')
		this.chatMenuItem = document.querySelector('#tab-menu-item-chat')
		this.recipesMenuItem = document.querySelector('#tab-menu-item-recipes')
		this.recipeButtons = [...document.querySelectorAll('.btn-recipe')]

		this.init()
	}

	init() {
		this.chatMenuItem.addEventListener('click', () => {
			this.setSelectedTab('chat')
		})
		this.recipesMenuItem.addEventListener('click', () => {
			this.setSelectedTab('recipes')
		})
		for (const recipeButton of this.recipeButtons) {
			recipeButton.addEventListener('click', event => {
				this.vscode.postMessage({ command: 'executeRecipe', recipe: event.target.dataset.recipe })
			})
		}
		this.renderTab()
	}

	setSelectedTab(newSelectedTab) {
		const changed = this.model.selectedTab !== newSelectedTab
		if (changed) {
			this.model.selectedTab = newSelectedTab
			this.renderTab()
		}
	}

	renderTab() {
		switch (this.model.selectedTab) {
			case 'chat':
				this.chatContainer.style.display = 'flex'
				this.recipesContainer.style.display = 'none'
				this.chatMenuItem.classList.add('tab-menu-item-selected')
				this.recipesMenuItem.classList.remove('tab-menu-item-selected')
				break
			case 'recipes':
				this.chatContainer.style.display = 'none'
				this.recipesContainer.style.display = 'flex'
				this.recipesMenuItem.classList.add('tab-menu-item-selected')
				this.chatMenuItem.classList.remove('tab-menu-item-selected')
				break
		}
	}
}

let controller = null

// TODO: We need a design for the chat empty state.
function onInitialize() {
	const vscode = acquireVsCodeApi()
	const inputElement = document.getElementById('input')
	const submitElement = document.querySelector('.submit-container')
	const resetElement = document.querySelector('.reset-conversation')

	const resizeInput = () => {
		inputElement.style.height = 0
		const height = Math.min(MAX_HEIGHT, inputElement.scrollHeight)
		inputElement.style.height = `${height}px`
		inputElement.style.overflowY = height >= MAX_HEIGHT ? 'auto' : 'hidden'
	}

	controller = new Controller(vscode, {
		selectedTab: 'chat',
	})

	inputElement.addEventListener('keydown', e => {
		if (e.key === 'Enter' && !e.shiftKey) {
			if (e.target.value.trim().length === 0) {
				return
			}
			vscode.postMessage({ command: 'submit', text: e.target.value })
			e.target.value = ''
			e.preventDefault()

			setTimeout(resizeInput, 0)
		}
	})

	inputElement.addEventListener('input', resizeInput)

	submitElement.addEventListener('click', () => {
		if (inputElement.value.trim().length === 0) {
			return
		}
		vscode.postMessage({ command: 'submit', text: inputElement.value })
		inputElement.value = ''
	})

	resetElement.addEventListener('click', () => {
		vscode.postMessage({ command: 'reset' })
	})

	vscode.postMessage({ command: 'initialized' })
}

function onMessage(event) {
	switch (event.data.type) {
		case 'transcript':
			renderMessages(event.data.messages, event.data.messageInProgress)
			break
		case 'showTab':
			if (controller) {
				controller.setSelectedTab(event.data.tab)
			}
			break
	}
}

const messageBubbleTemplate = `
<div class="bubble-row {type}-bubble-row">
	<div class="bubble {type}-bubble">
		<div class="bubble-content {type}-bubble-content">{text}</div>
		<div class="bubble-footer {type}-bubble-footer">
			{footer}
		</div>
	</div>
</div>
`

function getMessageBubble(author, text, timestamp) {
	const bubbleType = author === 'bot' ? 'bot' : 'human'
	const authorName = author === 'bot' ? 'Cody' : 'Me'
	return messageBubbleTemplate
		.replace(/{type}/g, bubbleType)
		.replace('{text}', text)
		.replace('{footer}', timestamp ? `${authorName} &middot; ${timestamp}` : `<i>${authorName} is writing...</i>`)
}

function getMessageInProgressBubble(author, text) {
	if (text.length === 0) {
		const loader = `
		<div class="bubble-loader">
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
			<div class="bubble-loader-dot"></div>
		</div>`
		return getMessageBubble(author, loader, null)
	}
	return getMessageBubble(author, text, null)
}

function renderMessages(messages, messageInProgress) {
	const inputElement = document.getElementById('input')
	const submitElement = document.querySelector('.submit-container')
	const transcriptContainerElement = document.querySelector('.transcript-container')

	const messageElements = messages
		.filter(message => !message.hidden)
		.map(message => getMessageBubble(message.speaker, message.displayText, message.timestamp))

	const messageInProgressElement = messageInProgress
		? getMessageInProgressBubble(messageInProgress.speaker, messageInProgress.displayText)
		: ''

	if (messageInProgress) {
		inputElement.setAttribute('disabled', '')
		submitElement.style.cursor = 'default'
	} else {
		inputElement.removeAttribute('disabled')
		submitElement.style.cursor = 'pointer'
	}

	transcriptContainerElement.innerHTML = messageElements.join('') + messageInProgressElement

	setTimeout(() => {
		if (messageInProgress && messageInProgress.displayText.length === 0) {
			document.querySelector('.bubble-loader')?.scrollIntoView()
		} else if (!messageInProgress && messages.length > 0) {
			document.querySelector('.bubble-row:last-child')?.scrollIntoView()
		}
	}, 0)
}

window.addEventListener('message', onMessage)
document.addEventListener('DOMContentLoaded', onInitialize)
