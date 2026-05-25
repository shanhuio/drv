import * as React from 'react' // for tsx

import * as apppage from '@shanhuio/misc/dist/apppage'

import * as dashcore from './dashcore'

export class TOTPSetup {
    SignedSecret: string = ''
    QRCode: string = ''
    URL: string = ''
}

export class TOTPData {
    Enabled: boolean = false
    TOTPSetup!: TOTPSetup
}

export class PageData {
    TOTP!: TOTPData
}

function sanityCheckOTP(code: string) {
    // Out OTP is explicitly configured to be 6 digits on the server side.
    if (!code) { return 'OTP is empty.' }
    if (!(code.length == 6 && /^\d+$/.test(code))) {
        return 'OTP should be a 6-digit number.'
    }
    return ''
}

export class EnableTotpPage {
    core: dashcore.Core

    enabled: boolean = false
    signedSecret: string = ''
    qr: string = ''
    errorMsg: string = ''
    otp: string = ''
    show: boolean = false

    constructor(core: dashcore.Core) {
        this.core = core
    }

    enter(path: string, pageData: any): apppage.Meta {
        this.core.setTab('2fa')
        this.core.fetchOrSet(this, path, pageData)
        return { title: 'Enable TOTP' }
    }

    exit() { this.clear() }

    redraw() { this.core.redraw() }

    setData(data: dashcore.PageData) {
        let d = (data.TwoFactorAuth as PageData).TOTP
        this.show = true
        this.enabled = d.Enabled
        this.errorMsg = ''
        if (this.enabled) {
            this.signedSecret = ''
            this.qr = ''
        } else {
            this.signedSecret = d.TOTPSetup.SignedSecret
            this.qr = d.TOTPSetup.QRCode
        }
    }

    clear() { this.show = false }

    submit() {
        this.errorMsg = sanityCheckOTP(this.otp)
        this.redraw()
        if (this.errorMsg) { return }

        this.core.app.call('/api/totp/enable', {
            SignedSecret: this.signedSecret,
            OTP: this.otp,
        }, {
            success: (resp: any, status: string, xhr: JQueryXHR) => {
                let msg = resp as { Error: string }
                if (msg.Error) {
                    this.errorMsg = msg.Error
                    this.redraw()
                    return
                }
                this.core.goto('2fa')
            },
            error: (xhr: JQueryXHR, status: string, err: string) => {
                this.errorMsg = 'Failed to enable TOTP: ' + xhr.responseText
                this.redraw()
            },
        })
    }

    renderError(): JSX.Element | null {
        if (!this.errorMsg) { return null }
        return <div className="error">
            <span className="error">{this.errorMsg}</span>
        </div>
    }

    render(): JSX.Element | null {
        if (!this.show) return null

        if (this.enabled) {
            return <div className="activate">
                TOTP already enabled.
            </div>
        }

        let onSubmit = (ev: React.FormEvent<HTMLElement>) => {
            ev.preventDefault()
            this.submit()
        }
        let onChange = (ev: React.FormEvent<HTMLInputElement>) => {
            this.otp = ev.currentTarget.value
        }
        let form = <form className="activate" onSubmit={onSubmit}>
            <div className="line">
                Open your TOTP Authentication App and
                scan the following QRCode:
            </div>
            <div className="qr"><img src={this.qr}></img></div>
            <div className="line">
                Enter the 6-digit code from your app:
            </div>
            <div className="line otp">
                <input type="text" onChange={onChange} />
            </div>
            {this.renderError()}
            <div className="submit">
                <input className="green-submit" type="submit"
                    value="Activate" />
            </div>
        </form>
        return <div className="two-factor">
            <h2>Enable TOTP</h2>
            {form}
        </div>
    }
}

export class DisableTotpPage {
    core: dashcore.Core
    errorMsg: string = ''
    show: boolean = false
    enabled: boolean = false

    constructor(core: dashcore.Core) {
        this.core = core
    }

    enter(path: string, data: any): apppage.Meta {
        this.core.setTab('2fa')
        this.core.fetchOrSet(this, path, data)
        return { title: 'Disable TOTP' }
    }

    exit() { this.clear() }

    clear() {
        this.show = false
        this.errorMsg = ''
    }

    setData(data: dashcore.PageData) {
        let d = data.TwoFactorAuth as PageData
        if (d.TOTP) { this.enabled = d.TOTP.Enabled }
        this.show = true
    }

    submit() {
        this.core.app.call(
            '/api/totp/disable', {}, {
            success: (resp: any, status: string, xhr: JQueryXHR) => {
                this.core.goto('2fa')
            },
            error: (xhr: JQueryXHR, status: string, err: string) => {
                this.errorMsg = 'Failed to disable TOTP: ' + xhr.responseText
                this.core.redraw()
            },
        })
    }

    renderError(): JSX.Element | null {
        if (!this.errorMsg) return null
        return <div className="error">
            <span className="error">{this.errorMsg}</span>
        </div>
    }

    render() {
        if (!this.show) return null

        let onClickGoBack = this.core.onClickGoto('2fa')

        if (!this.enabled) {
            return <React.Fragment>
                <h2>Disable TOTP</h2>
                <p>TOTP is already disabled.</p>
                <div className="buttons">
                    <a href="/2fa" onClick={onClickGoBack}>
                        <span className="button-gray">Go back</span>
                    </a>
                </div>
            </React.Fragment>
        }

        let onClickDisableConfirm = (ev: React.MouseEvent) => {
            ev.preventDefault()
            this.submit()
        }

        return <div className="two-factor">
            <h2>Disable TOTP</h2>
            <p>
                OTP authentication defends you from credential phishing
                and brute-force attacks. Are you sure you want to disable it?
            </p>
            <div className="buttons">
                <a href="/2fa" onClick={onClickGoBack}>
                    <span className="button-gray">Go back</span>
                </a>
                <a href="#" onClick={onClickDisableConfirm}>
                    <span className="button-red">Disable</span>
                </a>
            </div>
            {this.renderError()}
        </div>
    }
}

export class Page {
    core: dashcore.Core
    data: dashcore.PageData | null = null
    totpEnabled: boolean = false

    constructor(core: dashcore.Core) {
        this.core = core
    }

    enter(path: string, data: any): apppage.Meta {
        this.core.setTab('2fa')
        this.core.fetchOrSet(this, path, data)
        return { title: 'Two-factor Authentication' }
    }

    exit() { this.data = null }

    setData(data: dashcore.PageData) {
        this.data = data
        let d = data.TwoFactorAuth as PageData
        if (d.TOTP) { this.setTotpData(d.TOTP) }
    }

    setTotpData(d: TOTPData) {
        this.totpEnabled = d.Enabled
    }

    renderEnableButton(): JSX.Element {
        let onClick = this.core.onClickGoto('2fa/enable-totp')
        return <a href="/2fa/enable-totp" onClick={onClick}>
            <span className="button-green">Enable TOTP</span>
        </a>
    }

    renderDisableButton(): JSX.Element {
        let onClick = this.core.onClickGoto('2fa/disable-totp')
        return <a href="/2fa/disable-totp" onClick={onClick}>
            <span className="button-danger">Disable TOTP</span>
        </a>
    }

    renderState(): JSX.Element {
        if (this.totpEnabled) {
            return <div>
                <p>TOTP authentication is enabled.</p>
                {this.renderDisableButton()}
            </div>
        }
        return <div>
            <p>TOTP authentication is not enabled.</p>
            {this.renderEnableButton()}
        </div>
    }

    render(): JSX.Element | null {
        if (!this.data) return null
        return <div className="two-factor">
            <h2>Two-Factor Authentication</h2>
            {this.renderState()}
        </div>
    }
}
