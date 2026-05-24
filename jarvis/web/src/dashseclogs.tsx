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
import * as moment from '@shanhuio/misc/dist/moment'
import * as timeutil from '@shanhuio/misc/dist/timeutil'

import * as dashcore from './dashcore'

export class Entry {
    K: string = ''
    TSec: number = 0
    User: string = ''
    Text: string = ''
    Type: string = ''
    Vstr: string = ''
}

export class ViewData {
    entries: Entry[] = []
}

function renderList(data: Entry[], now: number) {
    if (!data) return <div className="empty">No entries.</div>

    let rows: JSX.Element[] = []
    for (let entry of data) {
        rows.push(<tr>
            <td>{moment.longFormat(now, entry.TSec)}</td>
            <td>{entry.Text}</td>
        </tr>)
    }

    return <table className="logs" cellPadding="0" cellSpacing="0">
        {rows}
    </table>
}

class PageData {
    Entries: Entry[] = []
}

export class Page {
    core: dashcore.Core
    now: timeutil.Timestamp = { Sec: 0, Nano: 0 }
    data: PageData | null = null

    constructor(c: dashcore.Core) {
        this.core = c
    }

    setData(data: dashcore.PageData) {
        this.data = data.SecurityLogs as PageData
        this.now = data.Now
    }

    enter(path: string, pageData: any): apppage.Meta {
        this.core.setTab('security-logs')
        this.core.fetchOrSet(this, path, pageData)
        return { title: 'Security Logs' }
    }

    exit() { this.data = null }

    render(): JSX.Element | null {
        if (!this.data) return null

        return <div className="logs">
            <h2>Security Logs</h2>
            {renderList(this.data.Entries, this.now.Sec)}
        </div>
    }
}
