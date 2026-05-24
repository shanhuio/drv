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

class DiskSize {
    MB: number = 0
    B: number = 0
}

function diskSizeSub(x: DiskSize, y: DiskSize): DiskSize {
    let mb = x.MB - y.MB
    let b = x.B - y.B
    if (b < 0) {
        b += 1e6
        mb -= 1
    }
    return { MB: mb, B: b }
}

class DiskUsage {
    Total!: DiskSize
    Free!: DiskSize
}

export class PageData {
    NoSysDock: boolean = false
    NextcloudDomain: string = ''
    IPAddrs: string[] = []
    DiskUsage: DiskUsage | null = null
}

function renderIPAddresses(d: PageData): JSX.Element | null {
    if (!d.IPAddrs) { return null }
    return <div>IP address: {d.IPAddrs.join(' ')}</div>
}

function renderDiskUsage(d: PageData): JSX.Element | null {
    let du = d.DiskUsage
    if (!du) return null

    let pp = (size: DiskSize) => {
        if (size.MB >= 1000) {
            // Size is in GBs.
            return (size.MB / 1000).toFixed(2) + 'GB'
        } else if (size.MB > 0) {
            // Size is in MBs.
            return (size.MB + size.B / 1e6).toFixed(2) + 'MB'
        }

        // Size is in KBs.
        return (size.B / 1000).toFixed(0) + 'KB'
    }

    let bytes = (size: DiskSize) => (size.MB * 1e6 + size.B)

    // Calculating percentage up to 2 decimal digits.
    let pct = (total: DiskSize, free: DiskSize) => {
        let totalBytes = bytes(total)
        if (totalBytes == 0) {
            console.log('total disk space should be greater than 0.')
            return 'unknown'
        }
        let freeBytes = bytes(free)
        return (100.0 * freeBytes / totalBytes).toFixed(2)
    }

    let used = diskSizeSub(du.Total, du.Free)
    return <p>
        Disk usage:
        Total {pp(du.Total)},
        Used {pp(used)} ({pct(du.Total, used)}%),
        Available {pp(du.Free)}
    </p>
}

function renderSystemStatus(d: PageData): JSX.Element {
    if (d.NoSysDock) {
        return <div className="section">
            <h3>System Status</h3>
            <p>Homedrive is not managing the operating system.</p>
        </div>
    }

    return <div className="section">
        <h3>System Status</h3>
        {renderIPAddresses(d)}
        {renderDiskUsage(d)}
    </div>
}

function renderApps(d: PageData): JSX.Element | null {
    let ncDomain = d.NextcloudDomain
    if (!ncDomain) return null
    return <div className="section">
        <h3>Applications</h3>
        <div className="apps">
            <div className="app">
                <a href={ncDomain}>
                    <img src="/img/nextcloud.png" className="logo" />
                </a>
            </div>
        </div>
    </div>
}

export class Page {
    core: dashcore.Core
    data: PageData | null = null

    constructor(core: dashcore.Core) {
        this.core = core
    }

    setData(d: dashcore.PageData) {
        this.data = d.Overview as PageData
    }

    enter(path: string, pageData: any): apppage.Meta {
        this.core.setTab('overview')
        this.core.fetchOrSet(this, path, pageData)
        return { title: 'Overview' }
    }

    exit() { this.data = null }

    render() {
        if (!this.data) { return null }
        return <div className="overview">
            {renderSystemStatus(this.data)}
            {renderApps(this.data)}
        </div>
    }
}
