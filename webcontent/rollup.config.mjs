import typescript from '@rollup/plugin-typescript';

export default {
    input: 'ts/comments.ts',
    output: {
        file: 'output/comments.js',
        format: 'iife',
    },
    plugins: [typescript({
        compilerOptions: {
            target: "esnext",
        }
    })],
}