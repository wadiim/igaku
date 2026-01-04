import { useState } from 'react';

interface Disease {
  id: string,
  rx_norm_id: string,
  name: string,
}

interface DiseaseTableProps {
  diseases: Disease[];
  page: number;
  totalPages: number;
  errorMessage: string | null;
  onPrev: () => void;
  onNext: () => void;
  onSelect: (disease: Disease) => void;
}

function DiseaseTable(
  { diseases, page, totalPages, errorMessage, onPrev, onNext, onSelect }: DiseaseTableProps
) {

  const [selectedIdx, setSelectedIdx] = useState<number | null>(null);

  const handleSelect = (idx: number) => {
    setSelectedIdx(idx);
    onSelect(diseases[idx]);
    console.log(diseases[idx]);
  };

  const listDiseases = diseases.map((disease, idx) =>
    <tr 
      key={disease.id}
      className={`bg-neutral-primary border-b border-default`}>
      <th 
        scope="row" 
        className={`
          px-6 py-4 
          font-medium text-heading whitespace-nowrap
        `}
      >
        {disease.name}
      </th>
      <td className={`px-6 py-4`}>
        {disease.rx_norm_id}
      </td>
      <td className={`px-6 py-4 text-center`}>
        <input
          type="radio"
          name="disease-select"
          checked={selectedIdx === idx}
          onChange={() => handleSelect(idx)}
          className={`
            border
            checked:border-brand
            focus:ring-brand-subtle 
          `}
        />
      </td>
    </tr>
  );

  return (
    <div className={`disease-pagination-table`}>
      <div 
        className={`
          relative 
          overflow-x-auto mb-2
          border
        `}
      >
        <table 
          className={`
            w-full text-sm text-left 
            rtl:text-right text-body`
          }
        >
          <thead className={`text-sm text-body border-b`}>
            <tr>
              <th scope="col" className={`px-6 py-3 font-medium`}>
                Name
              </th>
              <th scope="col" className={`px-6 py-3 font-medium`}>
                RxNorm ID
              </th>
              <th scope="col" className={`px-6 py-3 font-medium`}>
                Select
              </th>
            </tr>
          </thead>
          <tbody>
          {diseases.length === 0 ? (
            <tr>
              <td colSpan={4} className="px-6 py-4 text-center text-gray-500">
                No diseases found. Try a search.
              </td>
            </tr>
          ) : (
          listDiseases
          )}
          </tbody>
        </table>
      </div>
      {
        errorMessage &&
          <div className={`text-base text-tn-d-red mt-1 mb-1`}>{errorMessage}</div>
      }

      <div className={`flex flex-col items-center`}>
        <span className={`text-sm text-body`}>
            Showing <span>{page}</span> of <span>{totalPages}</span> pages
        </span>
        <div className={`inline-flex mt-4`}>
          <button 
            type="button" 
            onClick={onPrev}
            disabled={page <= 1}
            className={`
              inline-flex items-center border 
              font-medium text-sm px-4 py-2.5
            `}
          >
            Previous
          </button>
          <button 
            type="button" 
            onClick={onNext}
            disabled={page >= totalPages}
            className={`
              inline-flex items-center border 
              font-medium text-sm px-4 py-2.5
            `}
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
};

export default DiseaseTable;
