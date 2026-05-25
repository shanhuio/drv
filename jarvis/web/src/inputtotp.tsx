import * as React from 'react' // for tsx

import * as render from '@shanhuio/misc/dist/render'

export class PageData {
    SessionToken: string = ''
    Issuer: string = ''
    LoginError: string = ''
}

function renderError(error: string): JSX.Element | null {
    if (!error) { return null }
    return <div className="error">{error}</div>
}

function renderPage(data: PageData): JSX.Element {
    let form = <form className="login" action="/totp" method="post">
        <div className="line">
            <span className="prompt">TOTP</span>
            <input type="hidden" name="token" value={data.SessionToken} />
            <input className="password" type="text" autoFocus name="totp"
                autoComplete="off" />
            <input className="submit" type="submit" value="Verify" />
        </div>
        <div className="line hint">
            Input the code from your TOTP app for {data.Issuer}
        </div>
    </form>

    return <div className="login">
        <div className="logo"></div>
        {renderError(data.LoginError)}
        {form}
    </div>
}

export function main(data: PageData) {
    render.mainElement(renderPage(data))
}
