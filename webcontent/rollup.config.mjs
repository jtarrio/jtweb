import typescript from '@rollup/plugin-typescript';

export default [
    {
        input: 'ts/comments.ts',
        output: {
            file: 'generated/comments.js',
            format: 'iife',
        },
        plugins: [typescript({
            compilerOptions: {
                target: 'esnext',
            }
        })],
    },
    {
        input: 'ts/admin.ts',
        output: {
            file: 'generated/admin.js',
            format: 'iife',
        },
        plugins: [typescript({
            compilerOptions: {
                target: 'esnext',
            }
        })],
    }
]