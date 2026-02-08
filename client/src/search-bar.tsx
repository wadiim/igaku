import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons'

interface SearchBarProps {
  searchString: string;
  setSearchString: React.Dispatch<React.SetStateAction<string>>;
  onSearch: () => void;
}

function SearchBar({ searchString, setSearchString, onSearch }: SearchBarProps) {
  return (
    <form className="flex-auto">   
      <label htmlFor="search" 
        className={`
          block mb-2.5
          text-sm font-medium text-heading
          sr-only 
        `}
      >
        Search
      </label>
      <div className="relative">
        <div className={`absolute inset-y-0 flex items-center ps-1`}>
          <FontAwesomeIcon icon={faMagnifyingGlass} />
        </div>
        <input 
          type="search" 
          id="search" 
          className={`
            block w-full p-3 ps-9 pr-20
            border
            text-sm rounded-base 
          `}
          value={searchString}
          onChange={(e) => setSearchString(e.target.value)}
          placeholder="Search" required 
        />
        <button 
          type="button" 
          className={`
            absolute end-1.5 bottom-1.5 
            text-white bg-tn-d-dblue
            font-medium leading-5 rounded text-xs px-3 py-1.5
          `}
          onClick={() => onSearch()}
        >
          Search
        </button>
      </div>
    </form>
  );
}

export default SearchBar;
