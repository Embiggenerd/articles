/**
 * @type {import("../../types").ContainerHelper}
 */
const container = {
    id: 'container',
    type: 'div',
    classList: ['container'],
    root: 'root',
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        return element
    },
}

/**
 * @type {import("../../types").Render}
 */
const render = () => {
    const root = document.getElementById(container.root)
    if (!root) {
        throw new Error('no root to latch onto')
    }
    // Render chat area
    const containerElement = container.create()
    if (!containerElement) {
        throw new Error('failed to create container element')
    }
    root.append(containerElement)
    return {
        container: containerElement
    }
}

export default {
    render
}