import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <div
        className={`
          flex items-center justify-center h-screen
        `}
      >
        <button
          className={`
            bg-tn-d-black text-tn-d-fg
            px-4 py-2
            rounded-md
            cursor-pointer
            border border-transparent
            hover:border hover:border-tn-d-blue
            transition-border duration-250
          `}
          onClick={() => setCount((count) => count + 1)}
        >
          count is {count}
        </button>
      </div>
    </>
  )
}

export default App
