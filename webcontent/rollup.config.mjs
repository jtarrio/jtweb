import typescript from '@rollup/plugin-typescript';

export default {
    input: 'ts/comments.ts',
    output: {
        file: 'generated/comments.js',
        format: 'iife',
    },
    plugins: [typescript({
        compilerOptions: {
            target: "esnext",
        }
    })],
}