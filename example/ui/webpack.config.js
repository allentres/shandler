/* global require module __dirname process */
let path = require(`path`)
let webpack = require(`webpack`)
let HTTP_PORT = process.env.WEBPACK_DEV_PORT || 8888
let inProduction = (process.env.NODE_ENV === `production`)
let rules = [
	{
		test: /\.jsx?$/,
		use: [`babel-loader`],
		exclude: /node_modules/,
	}, {
		test: /\.css$/,
		use: [`style-loader`, `css-loader`],
	}, {
		test: /\.scss$/,
		use: [
			{
				loader: `style-loader`, // creates style nodes from JS strings
			}, {
				loader: `css-loader`, // translates CSS into CommonJS
			}, {
				loader: `sass-loader`, // compiles Sass to CSS
			},
		],
	}, {
		test: /\.(svg|png|gif|jpg|json)$/,
		use: [`file-loader`],
	}, {
		test: /\.(html|ico|zip|eot|svg)$/,
		loader: `file-loader?name=[name].[ext]`,
	}, {
		test: /\.(ttf|woff|woff2)$/,
		loader: `file-loader?name=fonts/[name].[ext]`,
	},
]

if (inProduction) {
	console.log(`Webpack Production mode`)
	module.exports = {
		entry: `./src/index.js`,
		output: {
			filename: `bundle.js`,
			path: path.resolve(__dirname, `dist`),
			publicPath: ``,
		},
		devtool: `source-map`,
		module: {rules},
		optimization: {
			minimize: true,
		},

	}
} else {
	console.log(`Webpack Development mode`)
	module.exports = {
		entry: `./src/index.js`,
		output: {
			filename: `bundle.js`,
			path: path.resolve(__dirname, `dist`),
			publicPath: ``,
		},
		devtool: `inline-source-map`,
		module: {rules},
		plugins: [
			new webpack.HotModuleReplacementPlugin(),
			new webpack.NamedModulesPlugin(),
			new webpack.NoEmitOnErrorsPlugin(),
		],
		devServer: {
			contentBase: __dirname + `/src`,
			compress: true,
			host: `localhost`,
			port: HTTP_PORT,
			https: false, // set to true to serve from https
			historyApiFallback: true, // respond to 404s with index.html
			hot: true, // enable HMR
			overlay: true,
			open: true,
			openPage: ``,
			quiet: true,
			noInfo: true,
		},
	}
}