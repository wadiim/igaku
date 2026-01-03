import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons'
import { isTokenExpired } from './utils/auth'

interface PatientData {
  username: string,
  email: string,
  national_id: string,
}

function Prescribe() {
  const [patientData, setPatientData] = useState<PatientData>({
    username: "",
    email: "",
    national_id: "",
  });

  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  let [searchString, setSearchString] = useState<string>("");

  let navigate = useNavigate();

  const onSearch = () => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/patient/${searchString}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      }) 
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid National ID"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        console.log(data);
        setPatientData(data); 
        setErrorMessage(null);
      })
      .catch((err) => {
        setErrorMessage(err.message);
      })
    }
  };

  return (
    <div className={`flex-1 flex flex-col items-center justify-center`}>
      <div
        className={`
          grid grid-cols-1
          text-tn-d-fg text-2xl
          border-2 border-tn-d-fg rounded-2x1 pb-0 p-4
        `}
      >
        <SearchBar 
            searchString={searchString}
            setSearchString={setSearchString}
            onSearch={onSearch}
        />
        {
          errorMessage &&
            <div className={`text-base text-tn-d-red mt-1 mb-1`}>{errorMessage}</div>
        }
        <div className={`border-1 mt-3 mb-3 pl-1 pr-1`}>
          <PatientItem title="Username" value={patientData.username} />
          <PatientItem title="Email" value={patientData.email} />
          <PatientItem title="National ID" value={patientData.national_id} />
        </div>
      </div>
    </div> 
  )
}

function PatientItem({ title, value }: { title: string, value: string}) {
  return (
    <div className={`grid grid-cols-2`}>
      <span className="font-bold">{ title }:</span>
      <div className={` overflow-x-auto whitespace-nowrap `}>
        <span>{ value }</span>
      </div>
    </div> 
  );
}

interface SearchBarProps {
  searchString: string;
  setSearchString: React.Dispatch<React.SetStateAction<string>>;
  onSearch: () => void;
}

function SearchBar({ searchString, setSearchString, onSearch}: SearchBarProps) {
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
          onClick={onSearch}
        >
          Search
        </button>
      </div>
    </form>
  );
}

export default Prescribe;
