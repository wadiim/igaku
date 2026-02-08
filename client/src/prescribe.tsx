import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'
import { sendNotification } from './utils/notify'
import SearchBar from './search-bar.tsx'
import DiseaseTable from './disease.tsx'
import DrugTable from './drug.tsx'

interface PatientData {
  id: string,
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

  const [diseaseData, setDiseaseData] = useState<Disease[]>([]);
  const [diseasePage, setDiseasePage] = useState<int>(1);
  const [diseaseTotalPages, setDiseaseTotalPages] = useState<int>(1);

  const [recDrugData, setRecDrugData] = useState<Drug[]>([]);
  const [recDrugPage, setRecDrugPage] = useState<int>(1);
  const [recDrugTotalPages, setRecDrugTotalPages] = useState<int>(1);

  const [manualDrugData, setManualDrugData] = useState<Drug[]>([]);
  const [manualDrugPage, setManualDrugPage] = useState<int>(1);
  const [manualDrugTotalPages, setManualDrugTotalPages] = useState<int>(1);

  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [diseaseErrorMessage, setDiseaseErrorMessage] = useState<string | null>(null);
  const [recDrugErrorMessage, setRecDrugErrorMessage] = useState<string | null>(null);
  const [manualDrugErrorMessage, setManualDrugErrorMessage] = useState<string | null>(null);

  let [patientSearchString, setPatientSearchString] = useState<string>("");
  let [diseaseSearchString, setDiseaseSearchString] = useState<string>("");
  let [drugSearchString, setDrugSearchString] = useState<string>("");

  let [selectedDisease, setSelectedDisease] = useState<Disease | null>(null);
  let [selectedDrugs, setSelectedDrugs] = useState<Drug[]>([]);

  const [recOrderBy, setRecOrderBy] = useState<"name" | "substance">("name");
  const [recOrderMethod, setRecOrderMethod] = useState<"asc" | "desc">("asc");

  const [manualOrderBy, setManualOrderBy] = useState<"name" | "substance">("name");
  const [manualOrderMethod, setManualOrderMethod] = useState<"asc" | "desc">("asc");

  const [selectedDrugIds, setSelectedDrugIds] = useState<Set<string>>(new Set());
  const [drugCache, setDrugCache] = useState<Map<string, Drug>>(new Map());

  let navigate = useNavigate();

  const handleDiseaseSelect = (disease: Disease) => {
    setSelectedDisease(disease); 
    recommendDrugs(disease, 1, recOrderBy, recOrderMethod);
  }

  const mergeIntoCache = (drugs: Drug[]) => {
    setDrugCache(prev => {
      const next = new Map(prev);
      drugs.forEach(d => next.set(d.id, d));
      return next;
    });
  };

  const handleToggleDrug = (drugId: string) => {
    setSelectedDrugIds(prev => {
      const next = new Set(prev);
      if (next.has(drugId)) {
        next.delete(drugId);
      } else {
        next.add(drugId);
      }
      return next;
    });
  };

  const toggleRecSort = (field: "name" | "substance") => {
    const newOrderBy = recOrderBy === field ? recOrderBy : field;
    const newOrderMethod =
      recOrderBy === field
        ? recOrderMethod === "asc"
        ? "desc"
        : "asc"
      : "asc";

    setRecOrderBy(newOrderBy);
    setRecOrderMethod(newOrderMethod);

    if (selectedDisease) {
      recommendDrugs(selectedDisease, 1, newOrderBy, newOrderMethod);
    }
  };

  const toggleManualSort = (field: "name" | "substance") => {
    const newOrderBy = manualOrderBy === field ? manualOrderBy : field;
    const newOrderMethod =
      manualOrderBy === field
        ? manualOrderMethod === "asc"
        ? "desc"
        : "asc"
      : "asc";

    setManualOrderBy(newOrderBy);
    setManualOrderMethod(newOrderMethod);

    onDrugSearch(manualDrugPage, newOrderBy, newOrderMethod);
  };

  const getSelectedDrugs = (): Drug[] => {
    return Array.from(selectedDrugIds)
      .map(id => drugCache.get(id))
      .filter((d): d is Drug => d !== undefined);
  };

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
      fetch(`http://localhost:4000/med/disease/${diseaseSearchString}?page=${page}`, {
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
        setDiseasePage(data.page);
        setDiseaseTotalPages(data.total_pages);
        setDiseaseErrorMessage(null);
      })
      .catch((err) => {
        setDiseaseErrorMessage(err.message);
        setDiseasePage(1);
      })
    }
  };

  const recommendDrugs = (
    disease: Disease,
    page: number = 1,
    orderBy: "name" | "substance",
    orderMethod: "asc" | "desc"
  ) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }
    if(jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/drug/recommend/${disease.rx_norm_id}?page=${page}&orderBy=${orderBy}&orderMethod=${orderMethod}`, {
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
          throw new Error("Drug not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setRecDrugData(data.data);
        setRecDrugPage(data.page);
        setRecDrugTotalPages(data.total_pages);
        setRecDrugErrorMessage(null);
        mergeIntoCache(data.data);
      })
      .catch((err) => { 
        setRecDrugData([]);
        setRecDrugPage(1);
        setRecDrugTotalPages(1);
        setRecDrugErrorMessage(err.message);
      })
    };
  }

  const onDrugSearch = (
    page: number = 1,
    orderBy: "name" | "substance",
    orderMethod: "asc" | "desc"
  ) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }
    if(jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/drug/${drugSearchString}?page=${page}&orderBy=${orderBy}&orderMethod=${orderMethod}`, {
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
          throw new Error("Drug not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setManualDrugData(data.data);
        setManualDrugPage(data.page);
        setManualDrugTotalPages(data.total_pages);
        mergeIntoCache(data.data);
      })
      .catch((err) => {
        setManualDrugErrorMessage(err.message);
      })
    }
  }

  const handleSubmit = () => {
    const payload = {
      patient: {
        id: patientData.id,
        username: patientData.username,
        email: patientData.email,
        national_id: patientData.national_id,
      },
      disease: {
        id: selectedDisease.id, 
        rx_norm_id: selectedDisease.rx_norm_id,
        name: selectedDisease.name,
      },
      drugs: getSelectedDrugs().map(d => ({
        id: d.id,
        name: d.name,
        substance: d.substance,
      })),
    };

    if (!payload.patient) {
      setErrorMessage("Patient must be selected before submitting.");
      return;
    }
    if (!payload.disease) {
      setErrorMessage("Select a disease first.");
      return;
    }
    if (payload.drugs.length === 0) {
      setErrorMessage("Choose at least one drug to prescribe.");
      return;
    }

    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if(jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/prescribe`, {
        method: "POST",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      })
      .then((res) => {
          sendNotification("Prescription submitted successfully");
          setSelectedDisease(null);
          setSelectedDrugIds(new Set());
          setRecDrugData([]);
          setManualDrugData([]);
          setErrorMessage(null);
      })
      .catch((err) => {
        console.log(err);
        setErrorMessage(err.message);
      })
      // TODO: Finish this endpoint
    }
  };

  return (
    <div className={`flex-1 flex items-center justify-center`}>
      <div
        className={`
          grid grid-cols-1
          text-tn-d-fg
          border-3 pb-0 p-4
          w-full max-w-4xl
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
            page={diseasePage} 
            totalPages={diseaseTotalPages} 
            errorMessage={diseaseErrorMessage}
            onPrev={() => {
              if (diseasePage > 1) {
                onDiseaseSearch(diseasePage - 1);
              }
            }}
            onNext={() => {
              if (diseasePage < diseaseTotalPages) {
                onDiseaseSearch(diseasePage + 1);
              }
            }}
            onSelect={handleDiseaseSelect}
          />
        </div>
        <h1 className={`text-2xl`}>Drugs</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <h2 className={`text-m`}>Recommended</h2>
          <DrugTable 
            tableName={"recommend"}
            drugs={recDrugData} 
            setDrugs={setRecDrugData}
            page={recDrugPage}
            totalPages={recDrugTotalPages}
            errorMessage={recDrugErrorMessage}
            onPrev={ () => {
              if (!selectedDisease) return;
              recommendDrugs(selectedDisease, recDrugPage - 1, recOrderBy, recOrderMethod);
            } }
            onNext={ () => {
              if (!selectedDisease) return;
              recommendDrugs(selectedDisease, recDrugPage + 1, recOrderBy, recOrderMethod);
            } }
            onSelect={handleToggleDrug}
            selectedIds={selectedDrugIds}
            orderBy={recOrderBy}
            orderMethod={recOrderMethod}
            onSortChange={toggleRecSort}
          />
          <h2 className={`text-m`}>Manual search</h2>
          <SearchBar 
              searchString={drugSearchString}
              setSearchString={setDrugSearchString}
              onSearch={ () => {onDrugSearch(1, manualOrderBy, manualOrderMethod)} }
          />
          <DrugTable 
            tableName={"manual"}
            drugs={manualDrugData}
            setDrugs={setManualDrugData}
            page={manualDrugPage}
            totalPages={manualDrugTotalPages}
            errorMessage={manualDrugErrorMessage}
            onPrev={ () => {onDrugSearch(manualDrugPage - 1, manualOrderBy, manualOrderMethod)} }
            onNext={ () => {onDrugSearch(manualDrugPage + 1, manualOrderBy, manualOrderMethod)} }
            onSelect={handleToggleDrug}
            selectedIds={selectedDrugIds}
            orderBy={manualOrderBy}
            orderMethod={manualOrderMethod}
            onSortChange={toggleManualSort}
          />
        </div>
        <div className="border-1 mt-6 p-4">
          <h2 className="text-xl mb-2">Summary</h2>

          <div className="mb-3">
            <ul>
              <li><strong>Username: </strong>{patientData.username}</li>  
              <li><strong>Email: </strong>{patientData.email}</li>  
              <li><strong>National ID: </strong>{patientData.national_id}</li>  
            </ul>
          </div>

          {selectedDisease && (
            <div className="mb-3">
              <strong>Disease:</strong> {selectedDisease.name}
            </div>
          )}

          <div className="mb-3">
            <strong>Drugs to prescribe:</strong>
            <ul className="list-disc list-inside">
              {getSelectedDrugs().map(d => (
                <li key={d.id}>{d.name} – {d.substance ?? ""}</li>
              ))}
            </ul>
          </div>

          <button
            onClick={handleSubmit}
            className={`
              px-4 py-2 bg-blue-600 text-white rounded
              hover:bg-blue-700 disabled:opacity-50
            `}
          >
            Submit
          </button>
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
