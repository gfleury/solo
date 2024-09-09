"use client";

import { MouseEvent } from "react";
import { Get, Delete, GetUrl, HandleFailures } from "../api-client";
import { Network } from "../api-client/models";
import Loading from "../loading";
import Table from "react-bootstrap/Table";
import Edit from "./edit";
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import { useSWRConfig } from "swr";
import { alertService } from "../services/alerts";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import SearchBar from "./search";
import { Stack } from "react-bootstrap";
import { faRemove } from "@fortawesome/free-solid-svg-icons";
import { OneOrBothBySize } from "../utils";
import { useRouter } from "next/navigation";

export default function Networks() {
  const { mutate } = useSWRConfig();
  const router = useRouter();
  const networks = HandleFailures(Get("/networks"), router);
  const networkDelete = Delete("/network/");

  function deleteFunction(networkID: number) {
    return async function deleteOnClick(e: MouseEvent<HTMLButtonElement>) {
      // @ts-ignore
      const data = await networkDelete.trigger(networkID);

      if (networkDelete.error || data.status != 204) {
        alertService.error(
          "Deleting network failed with: ".concat(await data.text()),
          {},
        );
      } else {
        mutate(GetUrl("/networks"));
      }
    };
  }

  function openNetwork(networkID: number) {
    return async function openOnClick(e: MouseEvent<HTMLTableDataCellElement>) {
      console.log(networkID);
      return router.push(`/nodes/network/${networkID}`);
    };
  }

  if (networks.isLoading) return Loading();

  const networksCast: Network[] = networks.data;

  return (
    <Stack className="my-2" gap={2}>
      <div>
        <SearchBar />
      </div>
      <div>
        <Table hover align="center" responsive>
          <thead>
            <tr>
              <th>Network</th>
              <th>CIDR</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {networksCast.map((network: Network) => (
              <tr key={network.ID}>
                <td
                  onClick={openNetwork(
                    network.ID !== undefined ? network.ID : -1,
                  )}
                >
                  {network.name}
                </td>
                <td
                  onClick={openNetwork(
                    network.ID !== undefined ? network.ID : -1,
                  )}
                >
                  {network.cidr}
                </td>
                <td>
                  <ButtonGroup size="sm">
                    <Button
                      variant="primary"
                      disabled={networkDelete.isMutating}
                      onClick={deleteFunction(
                        network.ID !== undefined ? network.ID : -1,
                      )}
                    >
                      {OneOrBothBySize(
                        "Delete",
                        <FontAwesomeIcon icon={faRemove} />,
                      )}
                    </Button>
                    <Edit
                      networkID={network.ID !== undefined ? network.ID : -1}
                    />
                  </ButtonGroup>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      </div>
    </Stack>
  );
}
