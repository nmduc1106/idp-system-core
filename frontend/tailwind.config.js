// tailwind.config.js
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  darkMode: 'class', // <--- Kiểm tra dòng này
  theme: {
    extend: {
      colors: {
        "primary": "#0078bd",
        "background-light": "#f5f7f8", // Màu xám nhạt của nền
        "background-dark": "#0f1c23",
      },
      keyframes: {
        indeterminate: {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(200%)' },
        }
      },
      animation: {
        indeterminate: 'indeterminate 1.5s infinite linear',
      }
    }
  }
}