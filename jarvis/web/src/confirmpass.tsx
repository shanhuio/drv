import * as React from 'react' // for tsx

import * as render from '@shanhuio/misc/dist/render'

export class PageData {
    RedirectTo: string = ''
    Error: string = ''
}

function renderError(error: string): JSX.Element | null {
    if (!error) { return null }
    return <div className="error">{error}</div>
}

function renderPage(data: PageData): JSX.Element {
    let form = <form action="/sudo" method="post">
        <div className="line">
            <span className="prompt">Confirm Password</span>
            <input className="password" type="password" autoFocus
                name="password" />
            <input className="password" type="hidden"
                name="redirect" value={data.RedirectTo} />
            <input className="submit" type="submit" value="Confirm" />
        </div>
    </form>

    return <div className="confirm-password">
        {renderError(data.Error)}
        {form}
    </div>
}

export function main(data: PageData) {
    render.mainElement(renderPage(data))
}
