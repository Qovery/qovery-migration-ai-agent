/** @type {import('next').NextConfig} */
const nextConfig = {
    webpack(config) {
        config.module.rules.push({
            test: /\.svg$/,
            use: ["@svgr/webpack"]
        });

        return config;
    },
    async rewrites() {
        return [
            {
                source: '/migrate/from/:paas/to/:iaas',
                destination: '/migrate/from/:paas/to/:iaas',
            },
        ];
    },
};

export default nextConfig;