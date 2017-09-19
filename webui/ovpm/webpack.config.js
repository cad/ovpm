const {resolve} = require('path');
const webpack = require('webpack');

module.exports = {
    context: resolve(__dirname, 'app'),

    entry: [

        'webpack-dev-server/client?http://localhost:8080',
        // bundle the client for webpack-dev-server
        // and connect to the provided endpoint

        './app.js'
        // the entry point of our app
    ],
    output: {
        filename: 'bundle.js',
        // the output bundle

        path: resolve(__dirname, 'public'),

    },

    devtool: 'inline-source-map',

    devServer: {

        contentBase: resolve(__dirname, 'public'),
        // match the output path

        publicPath: '/',
        // match the output `publicPath`

        historyApiFallback: true
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
    }
}
