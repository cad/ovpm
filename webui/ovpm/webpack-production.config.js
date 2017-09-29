const {resolve} = require('path');
const webpack = require('webpack');

module.exports = {
    context: resolve(__dirname, 'app'),

    entry: [

        './app.js'
        // the entry point of our app
    ],
    output: {
        filename: 'bundle.js',
        // the output bundle

        path: resolve(__dirname, 'public')
    },

    resolve: {
        extensions: [
            '.js', '.jsx', '.css', '.less'
        ],
        modules: ['./app', './node_modules']
    },

    module: {
        rules: [
            {
                test: /\.(js|jsx)?$/,
                use: ['babel-loader'],
                exclude: /node_modules/
            }, {
                test: /\.(less|css)$/,
                use: ['style-loader', 'css-loader?modules', 'less-loader']
            }, {
                test: /\.(jpg|png|gif)$/,
                use: 'file-loader'
            }, {
                test: /\.(woff|woff2|eot|ttf|svg)$/,
                use: {
                    loader: 'url-loader',
                    options: {
                        limit: 100000
                    }
                }
            }
        ]
    },

    plugins: [
        new webpack.LoaderOptionsPlugin({minimize: true, debug: false}),
        new webpack.DefinePlugin({
            'process.env': {
                'NODE_ENV': JSON.stringify('production')
            }
        }),
        new webpack.optimize.UglifyJsPlugin({
            beautify: false,
            mangle: {
                screw_ie8: true,
                keep_fnames: true
            },
            compress: {
                screw_ie8: true,
                warnings: false
            },
            comments: false
        })
    ],
    performance: {
        hints: "warning"
    }
};
