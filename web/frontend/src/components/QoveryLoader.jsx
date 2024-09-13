import React from 'react';
import {motion} from 'framer-motion';

const QoveryLoader = () => {
    return (
        <div className="flex justify-center items-center h-40">
            <motion.div
                animate={{
                    rotate: 360,
                    scale: [1, 1.2, 1],
                }}
                transition={{
                    duration: 2,
                    repeat: Infinity,
                    ease: "easeInOut"
                }}
            >
                <svg width="50" height="50" viewBox="0 0 50 50" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path
                        d="M25 0C11.1929 0 0 11.1929 0 25C0 38.8071 11.1929 50 25 50C38.8071 50 50 38.8071 50 25C50 11.1929 38.8071 0 25 0ZM25 45C13.9543 45 5 36.0457 5 25C5 13.9543 13.9543 5 25 5C36.0457 5 45 13.9543 45 25C45 36.0457 36.0457 45 25 45Z"
                        fill="#6453F7"/>
                    <path
                        d="M25 10C16.7157 10 10 16.7157 10 25C10 33.2843 16.7157 40 25 40C33.2843 40 40 33.2843 40 25C40 16.7157 33.2843 10 25 10ZM25 35C19.4772 35 15 30.5228 15 25C15 19.4772 19.4772 15 25 15C30.5228 15 35 19.4772 35 25C35 30.5228 30.5228 35 25 35Z"
                        fill="#6453F7"/>
                </svg>
            </motion.div>
        </div>
    );
};

export default QoveryLoader;