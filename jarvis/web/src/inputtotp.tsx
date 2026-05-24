// Copyright (C) 2022  Shanhu Tech Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
