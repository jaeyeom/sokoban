const path = require('path');
const webpack = require('webpack');

module.exports = {
  entry: './src/main.js',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  mode: 'development',
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-env']
          }
        }
      }
    ]
  },
  resolve: {
    alias: {
      phaser: path.resolve(__dirname, 'node_modules/phaser/dist/phaser.js')
    },
    fallback: {
      "fs": false,
      "path": false,
      "os": false,
      "buffer": require.resolve('buffer/'),
      "process": require.resolve('process/browser.js'), // Ensure .js extension is specified
      "stream": require.resolve('stream-browserify')
    },
    extensions: ['.js', '.mjs', '.json']
  },
  plugins: [
    new webpack.ProvidePlugin({
      Buffer: ['buffer', 'Buffer'],
      process: 'process/browser.js' // Ensure .js extension is specified
    })
  ]
};
