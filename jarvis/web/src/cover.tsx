import * as React from 'react' // for tsx

import * as redraw from '@shanhuio/misc/dist/redraw'
import * as render from '@shanhuio/misc/dist/render'

export class PageData {
    HideLogin: boolean = false
    RedirectTo: string = ''
    LoginError: string = ''
}

interface Props {
    pageData: PageData
}

class LoginFormConfig {
    loginError: string = ''
    redirectTo: string = ''
    hideForm: boolean = false
}

class LoginForm {
    error: string
    redraw: redraw.Redraw
    countDownSec: number = 0
    redirectStopped: boolean = false
    redirectTo: string = ''
    hideForm: boolean

    constructor(r: redraw.Redraw, config: LoginFormConfig) {
        this.redraw = r
        this.error = config.loginError
        this.redirectTo = config.redirectTo
        this.hideForm = config.hideForm
        if (this.redirectTo) {
            this.redirectStopped = false
            this.countDownSec = 5
        } else {
            this.redirectStopped = true
        }
    }

    startCountDown() {
        if (!this.redirectTo) return
        this.redirectStopped = false
        this.countDownSec = 5
        this.scheduleNextCoundDown()
    }

    scheduleNextCoundDown() {
        setTimeout(() => { this.countDown() }, 1000)
    }

    countDown() {
        if (this.redirectStopped) { return }
        if (this.countDownSec > 0) {
            this.countDownSec -= 1
            this.redraw()
            this.scheduleNextCoundDown()

            if (this.countDownSec <= 0) {
                window.location.replace(this.redirectTo)
            }
        }
    }

    renderError(): JSX.Element | null {
        if (!this.error) return null
        return <div className="error">{this.error}</div>
    }

    renderRedirectLink(): JSX.Element {
        return <a href={this.redirectTo}>Nextcloud</a>
    }

    renderRedirect(): JSX.Element | null {
        if (this.redirectStopped) return null
        if (this.countDownSec >= 2) {
            return <div className="redirect">
                Redirect to {this.renderRedirectLink()}
                {' in '} {this.countDownSec} seconds...
            </div>
        }
        if (this.countDownSec == 1) {
            return <div className="redirect">
                Redirect to {this.renderRedirectLink()} in
                1 second...
            </div>
        }
        return <div className="redirect">
            Redirect to {this.renderRedirectLink()} now...
        </div>
    }

    stopCountDown() {
        this.redirectStopped = true
        this.redraw()
    }

    renderForm(): JSX.Element | null {
        if (this.hideForm) return null
        return <form className="login" action="/login" method="post">
            <div className="line" onMouseDown={() => { this.stopCountDown() }} >
                <span className="prompt">Password</span>
                <input className="password" type="password" autoFocus
                    name="password"
                    onKeyDown={() => { this.stopCountDown() }}
                />
                <input className="submit" type="submit" value="Login" />
            </div>
        </form>
    }

    render(): JSX.Element {
        return <div className="login">
            <div className="logo"></div>
            {this.renderError()}
            {this.renderForm()}
            {this.renderRedirect()}
        </div>
    }
}

class Main extends React.Component<Props, {}> {
    redraw: redraw.Redraw
    loginForm: LoginForm
    data: PageData

    constructor(props: Props) {
        super(props)
        this.data = props.pageData
        this.redraw = redraw.NewRedraw(this)
        this.loginForm = new LoginForm(this.redraw, {
            hideForm: this.data.HideLogin,
            loginError: this.data.LoginError,
            redirectTo: this.data.RedirectTo,
        })
        this.loginForm.startCountDown()
    }

    render(): JSX.Element {
        return this.loginForm.render()
    }
}

export function main(data: PageData) {
    render.mainElement(<Main pageData={data} />)
}
