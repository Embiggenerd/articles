const appComponent = {
    renderer: null,
    mediaService: null,
    messageService: null,
    async init(renderer) {
        try {
            this.renderer = renderer.render()
            this.renderer.container.innerText = "renderer works"
        } catch (e) {
            this.handleError(e)
        }
    },

    
    handleError: function (error) {
        console.error(error)
        // const message = {
        //     text: error,
        //     from_user_name: 'ADMIN (to you)',
        // }
        // if (error.data) {
        //     message.text = error.data
        // }
        // if (error instanceof Error) {
        //     message.text = error.message
        // }
        // if (error.text) {
        //     message.text = error.text
        // }
        // this.renderer?.chatLog.addMessage(message)
    },

    
}

export default appComponent