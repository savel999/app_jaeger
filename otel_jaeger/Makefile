.PHONY: k8s_kustomization
k8s_kustomization:
	rm ./summary.yml
	kubectl kustomize ./k8s > ./summary.yml

.PHONY:k8s_kustomization k8s_apply
k8s_apply:
	kubectl apply -f ./summary.yml

.PHONY: k8s_kustomization k8s_delete
k8s_delete:
	kubectl delete -f ./summary.yml

################################################################################################################
