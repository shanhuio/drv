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

import * as apppage from '@shanhuio/misc/dist/apppage'

import * as dashcore from './dashcore'

export class PageData {
    Disabled: boolean = false
    Keys: string = ''
}

export class Page {
    core: dashcore.Core

    show: boolean = false
    disabled: boolean = false
    keys: string = ''
    successMsg: string = ''
    errorMsg: string = ''

    constructor(core: dashcore.Core) {
        this.core = core
    }

    redraw() { this.core.redraw() }

    update(ev: React.FormEvent<HTMLElement>) {
        ev.preventDefault()
        this.core.app.call('/api/sshkeys/update', {
            Keys: this.keys,
        }, {
            success: (resp: any, status: string, xhr: JQueryXHR) => {
                this.successMsg = 'Successfully updated SSH keys'
                this.redraw()
            },
            error: (xhr: JQueryXHR, status: string, err: string) => {
                console.log(err)
                this.errorMsg = 'Failed to update SSH keys: ' + err
                this.redraw()
            },
        })
    }

    enter(path: string, data: any): apppage.Meta {
        this.core.setTab('ssh-keys')
        this.core.fetchOrSet(this, path, data)
        return { title: 'SSH Keys' }
    }

    exit() { this.show = false }

    setData(data: dashcore.PageData) {
        this.show = true
        let d = data.SSHKeys as PageData
        this.disabled = d.Disabled
        this.keys = d.Keys
    }

    renderKeys(): JSX.Element {
        let onChange = (ev: React.FormEvent<HTMLTextAreaElement>) => {
            ev.preventDefault()
            const target = ev.target as HTMLTextAreaElement
            this.keys = target.value

            this.successMsg = ''
            this.errorMsg = ''

            this.redraw()
        }
        return <div>
            <textarea value={this.keys} onChange={onChange}></textarea>
        </div>
    }

    renderSuccess(): JSX.Element | null {
        if (!this.successMsg) return null
        return <div className="ok">{this.successMsg}</div>
    }

    renderError(): JSX.Element | null {
        if (!this.errorMsg) return null
        return <div className="error">{this.errorMsg}</div>
    }

    render(): JSX.Element | null {
        if (!this.show) return null

        let h2 = <h2>Authorized SSH Public Keys</h2>
        if (this.disabled) {
            return <React.Fragment>
                {h2}
                <p>HomeDrive is not managing the operating system, so
                    it does not manage the authorized SSH public keys. <br />
                    To change the SSH authorized keys, maybe change
                    them at <code>~/.ssh/authorized_keys</code>.</p>
            </React.Fragment>
        }
        let onClickUpdate = (ev: React.FormEvent<HTMLElement>) => {
            ev.preventDefault()
            this.update(ev)
            this.redraw()
        }
        return <div className="ssh-keys">
            {h2}
            {this.renderKeys()}
            <div className="controls"><a href="#" onClick={onClickUpdate}>
                <span className="button-green">
                    Update Keys
                </span>
            </a></div>
            {this.renderSuccess()}
            {this.renderError()}
        </div>
    }
}
