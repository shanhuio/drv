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

import * as redraw from '@shanhuio/misc/dist/redraw'
import * as render from '@shanhuio/misc/dist/render'
import * as appcore from '@shanhuio/misc/dist/appcore'
import * as apppage from '@shanhuio/misc/dist/apppage'

import * as dashpassword from './dashpassword'
import * as dashseclogs from './dashseclogs'
import * as dashcore from './dashcore'
import * as dash2fa from './dash2fa'
import * as dashsshkeys from './dashsshkeys'
import * as dashoverview from './dashoverview'

class Page {
    page: apppage.Page
    core: dashcore.Core

    constructor(core: dashcore.Core, page: apppage.Page) {
        this.page = page
        this.core = core
    }

    enter(path: string, pageData: any): apppage.Meta {
        this.core.redraw()
        return this.page.enter(path, pageData)
    }

    render() { return this.page.render() }
    exit() { return this.page.exit() }
}

class Dashboard {
    core: appcore.Core
    dashCore: dashcore.Core

    pages: Map<string, Page>

    constructor(r: redraw.Redraw, data: dashcore.PageData) {
        let s = new apppage.Switcher({ handler: this }, data.Path)
        this.core = appcore.make(r, s)
        this.dashCore = new dashcore.Core(this.core)

        let c = this.dashCore
        this.pages = new Map<string, Page>([
            ['overview', new Page(c, new dashoverview.Page(c))],
            ['change-password', new Page(c, new dashpassword.ChangePage(c))],
            ['security-logs', new Page(c, new dashseclogs.Page(c))],
            ['2fa', new Page(c, new dash2fa.Page(c))],
            ['2fa/enable-totp', new Page(c, new dash2fa.EnableTotpPage(c))],
            ['2fa/disable-totp', new Page(c, new dash2fa.DisableTotpPage(c))],
            ['ssh-keys', new Page(c, new dashsshkeys.Page(c))],
        ])

        s.init(data)
    }

    redraw() { this.core.redraw() }

    handle(path: string): apppage.Page {
        let p = this.pages.get(path)
        if (p) return p

        let overviewPage = this.pages.get('overview')
        if (overviewPage) return overviewPage

        throw new Error(`no page: ${path}`)
    }

    renderPage(): JSX.Element | null {
        let p = this.core.switcher.page()
        if (p) return p.render()
        return null
    }

    render(): JSX.Element {
        let tabs = [
            { name: 'ssh-keys', text: 'Authorized SSH Keys' },
            { name: 'security-logs', text: 'Security Logs' },
            { name: 'change-password', text: 'Change Password' },
            { name: '2fa', text: 'Two-Factor Authentication' },
        ]
        let tabLinks: JSX.Element[] = []
        for (let tab of tabs) {
            let url = '/' + tab.name
            let onClick = this.dashCore.onClickGoto(tab.name)
            tabLinks.push(<li>
                <a href={url} onClick={onClick}>{tab.text}</a>
            </li>)
        }

        let links = <ol className="links">{tabLinks}</ol>

        let onClickLogo = this.dashCore.onClickGoto('overview')

        return <React.Fragment>
            <div className="topbar">
                <span className="logo">
                    <a href="/overview" onClick={onClickLogo}>
                        <img className="logo" src='/img/logo-white.png' />
                    </a>
                </span>
                <a href="/overview" onClick={onClickLogo}>HomeDrive</a>
                <span className="right">
                    <span className="part">
                        <a href="/signout">Sign Out</a>
                    </span>
                </span>
            </div>
            <div className="menu">{links}</div>
            <div className="view">{this.renderPage()}</div>
        </React.Fragment>
    }
}

interface Props {
    data: dashcore.PageData
}

class Main extends React.Component<Props, {}> {
    redraw: redraw.Redraw
    dashboard: Dashboard

    constructor(props: Props) {
        super(props)
        this.redraw = redraw.NewRedraw(this)
        this.dashboard = new Dashboard(this.redraw, props.data)
    }

    render() { return this.dashboard.render() }
}

export function main(data: dashcore.PageData) {
    render.mainElement(<Main data={data}></Main>)
}
