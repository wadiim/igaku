import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'
import SearchBar from './search-bar.tsx'
import DiseaseTable from './disease.tsx'

interface PatientData {
  username: string,
  email: string,
  national_id: string,
}

interface Disease {
  id: string,
  rx_norm_id: string,
  name: string,
}

function Prescribe() {
  const [patientData, setPatientData] = useState<PatientData>({
    username: "",
    email: "",
    national_id: "",
  });

  const PAGE_SIZE = 5;

  const [diseaseData, setDiseaseData] = useState<Disease[]>([]);
  const [pageNumber, setPageNumber] = useState<int>(1);
  const [totalPageNumber, setTotalPageNumber] = useState<int>(1);

  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [diseaseErrorMessage, setDiseaseErrorMessage] = useState<string | null>(null);

  let [patientSearchString, setPatientSearchString] = useState<string>("");
  let [diseaseSearchString, setDiseaseSearchString] = useState<string>("");

  let [selectedDisease, setSelectedDisease] = useState<Disease | null>(null);

  let navigate = useNavigate();

  const handleDiseaseSelect = (disease: Disease) => {
    setSelectedDisease(disease); 
  }

  const onPatientSearch = () => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/patient/${patientSearchString}`, {
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

  const onDiseaseSearch = (page = 1) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/disease/${diseaseSearchString}?page=${page}&pageSize=${PAGE_SIZE}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      }) 
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid parameters"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 404) {
          throw new Error("Disease not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setDiseaseData(data.data);
        setPageNumber(data.page);
        setTotalPageNumber(data.total_pages);
        setDiseaseErrorMessage(null);
      })
      .catch((err) => {
        setDiseaseErrorMessage(err.message);
        setPageNumber(1);
      })
    }
  };

  return (
    <div className={`flex-1 flex flex-col items-center justify-center`}>
      <div
        className={`
          grid grid-cols-1
          text-tn-d-fg
          border-3 pb-0 p-4
        `}
      >
        <h1 className={`text-2xl`}>Patient</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <SearchBar 
              searchString={patientSearchString}
              setSearchString={setPatientSearchString}
              onSearch={onPatientSearch}
          />
          {
            errorMessage &&
              <div className={`text-base text-tn-d-red mt-1 mb-1`}>{errorMessage}</div>
          }
          <PatientDetails title="Username" value={patientData.username} />
          <PatientDetails title="Email" value={patientData.email} />
          <PatientDetails title="National ID" value={patientData.national_id} />
        </div>
        <h1 className={`text-2xl`}>Disease</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <SearchBar 
              searchString={diseaseSearchString}
              setSearchString={setDiseaseSearchString}
              onSearch={onDiseaseSearch}
          />
          <DiseaseTable 
            diseases={diseaseData} 
            page={pageNumber} 
            totalPages={totalPageNumber} 
            errorMessage={diseaseErrorMessage}
            onPrev={() => {
              if (pageNumber > 1) {
                onDiseaseSearch(pageNumber - 1);
              }
            }}
            onNext={() => {
              if (pageNumber < totalPageNumber) {
                onDiseaseSearch(pageNumber + 1);
              }
            }}
            onSelect={handleDiseaseSelect}
          />
        </div>

      </div>
    </div> 
  )
}

function PatientDetails({ title, value }: { title: string, value: string}) {
  return (
    <div className={`grid grid-cols-2`}>
      <span className={`font-medium text-sm pt-2`}>{ title }:</span>
      <div className={`overflow-x-auto whitespace-nowrap`}>
        <span>{ value }</span>
      </div>
    </div> 
  );
}

export default Prescribe;
