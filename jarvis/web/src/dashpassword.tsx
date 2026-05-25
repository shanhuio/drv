import * as React from 'react' // for tsx

import * as inputs from '@shanhuio/misc/dist/inputs'
import * as apppage from '@shanhuio/misc/dist/apppage'

import * as dashcore from './dashcore'

export class ChangePage {
    core: dashcore.Core

    session: number = 0 // For redrawing the input boxes.

    oldPassword: string = ''
    newPassword: string = ''
    newRepeat: string = ''
    error: string = ''
    ok: boolean = false

    constructor(core: dashcore.Core) {
        this.clear()
        this.core = core
    }

    enter(path: string, data: any): apppage.Meta {
        this.core.setTab('change-password')
        return { title: 'Change Password' }
    }

    exit() { this.clear() }

    clear() {
        this.session += 1
        this.oldPassword = ''
        this.newPassword = ''
        this.newRepeat = ''
        this.error = ''
        this.ok = false
    }

    checkInput() {
        if (this.oldPassword == '') return 'Old password cannot be empty.'
        if (this.newPassword == '') return 'New password cannot be empty.'
        if (this.newRepeat == '') return 'Please confirm the password again.'
        if (this.newPassword != this.newRepeat) {
            return 'New passwords do not match.'
        }
        if (this.newPassword == this.oldPassword) {
            return 'The new password is the same as the old one.'
        }
        return ''
    }

    submit(ev: React.FormEvent<HTMLElement>) {
        ev.preventDefault()

        let err = this.checkInput()
        this.error = err
        this.ok = false
        this.core.redraw()
        if (err) return

        this.core.app.call('/api/user/changepwd', {
            OldPassword: this.oldPassword,
            NewPassword: this.newPassword,
        }, {
            success: (resp: any, status: string, xhr: JQueryXHR) => {
                let r = resp as { Error: string }
                if (r.Error) {
                    this.error = r.Error
                    this.core.redraw()
                    return
                }

                this.clear()
                this.ok = true
                this.core.redraw()
            },
            error: (xhr: JQueryXHR, status: string, err: string) => {
                if (xhr.status == 403) {
                    this.error = xhr.responseText
                    this.core.redraw()
                    return
                }
                console.log(err)
                this.error = 'Set new password failed.'
                this.core.redraw()
            },
        })
    }

    renderError(): JSX.Element | null {
        if (!this.error) return null
        return <span className="warning">{this.error}</span>
    }

    renderOK(): JSX.Element | null {
        if (this.error) return null
        if (!this.ok) return null
        return <span className="ok">Password changed.</span>
    }

    render(): JSX.Element {
        let form = <form className="password" onSubmit={
            (ev: React.FormEvent<HTMLElement>) => { this.submit(ev) }}>
            <div className="line">
                <span className="field">Current password</span>
                <inputs.Password
                    autoFocus focus session={this.session}
                    onChange={(v: string) => { this.oldPassword = v }}
                />
            </div>
            <div className="line">
                <span className="field">New password</span>
                <inputs.Password
                    session={this.session}
                    onChange={(v: string) => { this.newPassword = v }}
                />
            </div>
            <div className="line">
                <span className="field">Type again</span>
                <inputs.Password session={this.session}
                    onChange={(v: string) => { this.newRepeat = v }}
                />
            </div>
            <div className="submit">
                <input className="submit" type="submit"
                    value="Change" />
                {this.renderError()}
                {this.renderOK()}
            </div>
        </form>
        return <div className="password">
            <h2>Change Password</h2>
            {form}
        </div>
    }
}
