"use strict";

var esbuild = require('esbuild')

var entryPoints = [
    './src/mains/dashboard.tsx',
    './src/mains/cover.tsx',
    './src/mains/confirmpass.tsx',
    './src/mains/inputtotp.tsx',
]

esbuild.buildSync({
    entryPoints: entryPoints,
    bundle: true,
    define: {
        'process.env.NODE_ENV': '"production"',
    },
    minify: true,
    sourcemap: true,
    outdir: './dist/webdist',
    external: ['react-dom'],
})
